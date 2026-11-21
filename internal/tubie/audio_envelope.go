package tubie

import (
	"encoding/binary"
	"math"
	"os"
	"sort"
)

type wavPCMData struct {
	sampleRate int
	samples    []float64
}

func readWAVMonoSamples(wavPath string) (*wavPCMData, error) {
	data, err := os.ReadFile(wavPath)
	if err != nil {
		return nil, err
	}
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return nil, os.ErrInvalid
	}

	var (
		hasFmt        bool
		audioFormat   uint16
		numChannels   uint16
		sampleRate    uint32
		bitsPerSample uint16
		pcmBytes      []byte
	)

	offset := 12
	for offset+8 <= len(data) {
		chunkID := string(data[offset : offset+4])
		chunkSize := int(binary.LittleEndian.Uint32(data[offset+4 : offset+8]))
		payloadStart := offset + 8
		payloadEnd := payloadStart + chunkSize
		if chunkSize < 0 || payloadEnd > len(data) {
			break
		}

		switch chunkID {
		case "fmt ":
			if chunkSize >= 16 {
				audioFormat = binary.LittleEndian.Uint16(data[payloadStart : payloadStart+2])
				numChannels = binary.LittleEndian.Uint16(data[payloadStart+2 : payloadStart+4])
				sampleRate = binary.LittleEndian.Uint32(data[payloadStart+4 : payloadStart+8])
				bitsPerSample = binary.LittleEndian.Uint16(data[payloadStart+14 : payloadStart+16])
				if audioFormat == 0xFFFE && chunkSize >= 40 {
					audioFormat = binary.LittleEndian.Uint16(data[payloadStart+24 : payloadStart+26])
				}
				hasFmt = true
			}
		case "data":
			pcmBytes = data[payloadStart:payloadEnd]
		}

		offset = payloadEnd
		if chunkSize%2 == 1 {
			offset++
		}
		if hasFmt && pcmBytes != nil {
			break
		}
	}

	if !hasFmt || len(pcmBytes) == 0 || sampleRate == 0 || numChannels == 0 {
		return nil, os.ErrInvalid
	}

	bytesPerSample := int(bitsPerSample) / 8
	if bytesPerSample <= 0 {
		return nil, os.ErrInvalid
	}
	frameBytes := bytesPerSample * int(numChannels)
	if frameBytes <= 0 {
		return nil, os.ErrInvalid
	}
	numFrames := len(pcmBytes) / frameBytes
	if numFrames == 0 {
		return nil, os.ErrInvalid
	}

	mono := make([]float64, numFrames)
	chCount := int(numChannels)

	switch {
	case audioFormat == 1 && bitsPerSample == 16:
		for i := 0; i < numFrames; i++ {
			base := i * frameBytes
			var sum float64
			for c := 0; c < chCount; c++ {
				pos := base + c*2
				v := int16(binary.LittleEndian.Uint16(pcmBytes[pos : pos+2]))
				sum += float64(v) / 32768.0
			}
			mono[i] = sum / float64(chCount)
		}
	case audioFormat == 1 && bitsPerSample == 8:
		for i := 0; i < numFrames; i++ {
			base := i * frameBytes
			var sum float64
			for c := 0; c < chCount; c++ {
				v := float64(pcmBytes[base+c]) - 128.0
				sum += v / 128.0
			}
			mono[i] = sum / float64(chCount)
		}
	case audioFormat == 3 && bitsPerSample == 32:
		for i := 0; i < numFrames; i++ {
			base := i * frameBytes
			var sum float64
			for c := 0; c < chCount; c++ {
				pos := base + c*4
				bits := binary.LittleEndian.Uint32(pcmBytes[pos : pos+4])
				sum += float64(math.Float32frombits(bits))
			}
			mono[i] = sum / float64(chCount)
		}
	default:
		return nil, os.ErrInvalid
	}

	return &wavPCMData{
		sampleRate: int(sampleRate),
		samples:    mono,
	}, nil
}

// ExtractSpeechLevels computes normalized per-frame RMS speech intensity in [0, 1] directly from a WAV file.
// Returns nil if the WAV file is missing, unreadable, or in an unsupported encoding.
func ExtractSpeechLevels(wavPath string, fps int, duration float64) []float64 {
	if wavPath == "" {
		return nil
	}
	wav, err := readWAVMonoSamples(wavPath)
	if err != nil || len(wav.samples) == 0 || wav.sampleRate <= 0 {
		return nil
	}
	if fps <= 0 {
		fps = 30
	}
	if duration <= 0 {
		duration = float64(len(wav.samples)) / float64(wav.sampleRate)
	}
	totalFrames := int(math.Ceil(duration * float64(fps)))
	if totalFrames <= 0 {
		return nil
	}

	rms := make([]float64, totalFrames)
	maxRMS := 0.0
	sr := float64(wav.sampleRate)
	fpsF := float64(fps)
	nSamples := len(wav.samples)

	for f := 0; f < totalFrames; f++ {
		start := int(math.Floor(float64(f) * sr / fpsF))
		end := int(math.Floor(float64(f+1) * sr / fpsF))
		if start < 0 {
			start = 0
		}
		if start >= nSamples {
			continue
		}
		if end <= start {
			end = start + 1
		}
		if end > nSamples {
			end = nSamples
		}
		var sumSq float64
		for i := start; i < end; i++ {
			s := wav.samples[i]
			sumSq += s * s
		}
		val := math.Sqrt(sumSq / float64(end-start))
		rms[f] = val
		if val > maxRMS {
			maxRMS = val
		}
	}

	sorted := make([]float64, len(rms))
	copy(sorted, rms)
	sort.Float64s(sorted)
	p95Idx := int(math.Floor(0.95 * float64(len(sorted)-1)))
	ref := sorted[p95Idx]
	if ref < 1e-5 {
		if maxRMS >= 1e-5 {
			ref = maxRMS
		} else {
			ref = 1e-5
		}
	}

	const attack = 0.58
	const release = 0.28
	levels := make([]float64, totalFrames)
	prev := 0.0
	for i, r := range rms {
		norm := math.Min(1.0, r/ref)
		coeff := release
		if norm > prev {
			coeff = attack
		}
		smoothed := prev + coeff*(norm-prev)
		if smoothed < 0.005 {
			smoothed = 0
		}
		smoothed = math.Max(0, math.Min(1, smoothed))
		levels[i] = math.Round(smoothed*100) / 100
		prev = smoothed
	}
	return levels
}

// DetectSpeechSegmentsFromWAV finds voiced speech regions directly from a WAV file without external tools.
func DetectSpeechSegmentsFromWAV(wavPath string, totalDuration float64) []SpeechSegment {
	if wavPath == "" {
		return nil
	}
	wav, err := readWAVMonoSamples(wavPath)
	if err != nil || len(wav.samples) == 0 || wav.sampleRate <= 0 {
		return nil
	}
	if totalDuration <= 0 {
		totalDuration = float64(len(wav.samples)) / float64(wav.sampleRate)
	}

	frameSamples := wav.sampleRate / 50 // 20ms buckets
	if frameSamples < 1 {
		frameSamples = 1
	}
	numFrames := (len(wav.samples) + frameSamples - 1) / frameSamples
	if numFrames == 0 {
		return nil
	}

	frameRMS := make([]float64, numFrames)
	for f := 0; f < numFrames; f++ {
		start := f * frameSamples
		end := start + frameSamples
		if end > len(wav.samples) {
			end = len(wav.samples)
		}
		var sumSq float64
		for i := start; i < end; i++ {
			s := wav.samples[i]
			sumSq += s * s
		}
		frameRMS[f] = math.Sqrt(sumSq / float64(end-start))
	}

	sorted := make([]float64, len(frameRMS))
	copy(sorted, frameRMS)
	sort.Float64s(sorted)
	p90 := sorted[int(math.Floor(0.90*float64(len(sorted)-1)))]
	thr := math.Max(0.015, p90*0.08)

	raw := make([]SpeechSegment, 0)
	inSpeech := false
	segStart := 0.0
	sr := float64(wav.sampleRate)

	for f, r := range frameRMS {
		tStart := float64(f*frameSamples) / sr
		tEnd := math.Min(totalDuration, float64((f+1)*frameSamples)/sr)
		if r >= thr {
			if !inSpeech {
				segStart = tStart
				inSpeech = true
			}
		} else if inSpeech {
			raw = append(raw, SpeechSegment{Start: segStart, End: tStart})
			inSpeech = false
		}
		if f == len(frameRMS)-1 && inSpeech {
			raw = append(raw, SpeechSegment{Start: segStart, End: tEnd})
		}
	}

	merged := make([]SpeechSegment, 0, len(raw))
	for _, seg := range raw {
		st := math.Max(0, math.Min(seg.Start, totalDuration))
		en := math.Max(0, math.Min(seg.End, totalDuration))
		if en <= st {
			continue
		}
		if len(merged) > 0 && st-merged[len(merged)-1].End < 0.14 {
			merged[len(merged)-1].End = math.Max(merged[len(merged)-1].End, en)
		} else {
			merged = append(merged, SpeechSegment{Start: st, End: en})
		}
	}

	out := make([]SpeechSegment, 0, len(merged))
	for _, seg := range merged {
		if seg.End-seg.Start >= 0.05 {
			out = append(out, seg)
		}
	}
	return out
}
