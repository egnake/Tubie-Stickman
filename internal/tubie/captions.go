package tubie

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"unicode"
)

// TimedWord represents an individual spoken word with millisecond precision.
type TimedWord struct {
	Word  string  `json:"word"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

// CaptionCue represents a 1-3 word animated viral caption block.
type CaptionCue struct {
	Start float64     `json:"start"`
	End   float64     `json:"end"`
	Text  string      `json:"text"`
	Words []TimedWord `json:"words"`
}

// CleanTurkishWord cleans punctuation and converts to uppercase with Turkish casing.
func CleanTurkishWord(w string) string {
	w = strings.TrimSpace(w)
	re := regexp.MustCompile(`[^\p{L}\p{N}-]+`)
	cleaned := re.ReplaceAllString(w, "")
	return strings.ToUpperSpecial(unicode.TurkishCase, cleaned)
}

// ExtractScriptWords breaks script text into clean Turkish uppercase words.
func ExtractScriptWords(scriptText string) []string {
	fields := strings.Fields(scriptText)
	words := make([]string, 0, len(fields))
	for _, f := range fields {
		clean := CleanTurkishWord(f)
		if clean != "" {
			words = append(words, clean)
		}
	}
	return words
}

// ExtractWhisperTimestamps runs faster-whisper to extract acoustic timing intervals.
func ExtractWhisperTimestamps(wavPath string) ([]map[string]float64, error) {
	pyCode := fmt.Sprintf(`
import json, sys
try:
    from faster_whisper import WhisperModel
    model = WhisperModel("base", device="cpu", compute_type="int8")
    segments, _ = model.transcribe(%q, language="tr", word_timestamps=True)
    intervals = []
    for s in segments:
        for w in s.words:
            intervals.append({"start": round(w.start, 3), "end": round(w.end, 3)})
    print(json.dumps(intervals))
except Exception as e:
    sys.exit(1)
`, wavPath)

	cmd := exec.Command(BinaryPath("python"), "-c", pyCode)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var intervals []map[string]float64
	if err := json.Unmarshal(out, &intervals); err != nil {
		return nil, err
	}
	return intervals, nil
}

// BuildAccurateCaptions creates perfectly spelled, aligned viral captions from the script.
func BuildAccurateCaptions(scriptText string, wavPath string, totalDuration float64) []CaptionCue {
	canonicalWords := ExtractScriptWords(scriptText)
	if len(canonicalWords) == 0 {
		return []CaptionCue{}
	}

	// 1. Try to get acoustic timings
	var intervals []map[string]float64
	if wavPath != "" {
		if fi, err := os.Stat(wavPath); err == nil && !fi.IsDir() && fi.Size() > 0 {
			if wIntervals, err := ExtractWhisperTimestamps(wavPath); err == nil && len(wIntervals) > 0 {
				intervals = wIntervals
			}
		}
	}

	// 2. Map canonical words to timings
	timedWords := make([]TimedWord, 0, len(canonicalWords))
	totalWords := len(canonicalWords)

	if len(intervals) >= totalWords {
		// We have enough acoustic intervals: map directly 1-to-1
		for i, word := range canonicalWords {
			st := intervals[i]["start"]
			en := intervals[i]["end"]
			if en <= st {
				en = st + 0.25
			}
			timedWords = append(timedWords, TimedWord{
				Word:  word,
				Start: math.Round(st*1000) / 1000,
				End:   math.Round(en*1000) / 1000,
			})
		}
	} else if len(intervals) > 0 {
		// Scale intervals proportionally to cover all canonical words
		ratio := float64(len(intervals)) / float64(totalWords)
		for i, word := range canonicalWords {
			idx := int(float64(i) * ratio)
			if idx >= len(intervals) {
				idx = len(intervals) - 1
			}
			st := intervals[idx]["start"]
			en := intervals[idx]["end"]
			if en <= st {
				en = st + 0.25
			}
			timedWords = append(timedWords, TimedWord{
				Word:  word,
				Start: math.Round(st*1000) / 1000,
				End:   math.Round(en*1000) / 1000,
			})
		}
	} else {
		// Proportional distribution across total duration
		lead := 0.2
		usable := totalDuration - lead - 0.4
		if usable <= 0 {
			usable = totalDuration
			lead = 0
		}
		step := usable / float64(totalWords)
		for i, word := range canonicalWords {
			st := lead + float64(i)*step
			en := st + step*0.85
			timedWords = append(timedWords, TimedWord{
				Word:  word,
				Start: math.Round(st*1000) / 1000,
				End:   math.Round(en*1000) / 1000,
			})
		}
	}

	// 3. Group into 1-3 word cues (Alex Hormozi viral rhythm)
	rawCues := make([]CaptionCue, 0)
	i := 0
	n := len(timedWords)

	for i < n {
		take := 2
		if i+3 <= n {
			len3 := len(timedWords[i].Word) + len(timedWords[i+1].Word) + len(timedWords[i+2].Word)
			if len3 <= 16 {
				take = 3
			} else if len(timedWords[i].Word) > 8 {
				take = 1
			}
		} else if i+1 == n {
			take = 1
		}

		chunk := timedWords[i : i+take]
		textParts := make([]string, len(chunk))
		for k, w := range chunk {
			textParts[k] = w.Word
		}

		rawCues = append(rawCues, CaptionCue{
			Start: chunk[0].Start,
			End:   chunk[len(chunk)-1].End,
			Text:  strings.Join(textParts, " "),
			Words: chunk,
		})
		i += take
	}

	// 4. Strict Sanitization: Guarantee Monotonic, Non-overlapping, In-bounds
	sanitized := make([]CaptionCue, 0, len(rawCues))
	cueEnd := 0.0

	for _, cue := range rawCues {
		st := cue.Start
		if st < cueEnd {
			st = cueEnd
		}
		en := cue.End
		if en < st+0.05 {
			en = st + 0.05
		}
		if en > totalDuration {
			en = totalDuration
		}
		if st >= totalDuration-0.02 {
			break
		}
		if en <= st {
			en = math.Min(st+0.05, totalDuration)
			if en <= st {
				break
			}
		}

		st = math.Round(st*1000) / 1000
		en = math.Round(en*1000) / 1000

		// Re-clamp inside words
		words := make([]TimedWord, len(cue.Words))
		wCount := len(cue.Words)
		wStep := (en - st) / float64(wCount)
		for wi, w := range cue.Words {
			wSt := math.Round((st+float64(wi)*wStep)*1000) / 1000
			wEn := math.Round((st+float64(wi+1)*wStep)*1000) / 1000
			words[wi] = TimedWord{
				Word:  w.Word,
				Start: wSt,
				End:   wEn,
			}
		}

		sanitized = append(sanitized, CaptionCue{
			Start: st,
			End:   en,
			Text:  cue.Text,
			Words: words,
		})
		cueEnd = en
	}

	return sanitized
}

// FormatSRT converts caption cues into standard SubRip (.srt) subtitle format.
func FormatSRT(cues []CaptionCue) string {
	var sb strings.Builder
	for i, c := range cues {
		stH := int(c.Start / 3600)
		stM := int(math.Mod(c.Start, 3600) / 60)
		stS := int(math.Mod(c.Start, 60))
		stMs := int(math.Mod(c.Start*1000, 1000))

		enH := int(c.End / 3600)
		enM := int(math.Mod(c.End, 3600) / 60)
		enS := int(math.Mod(c.End, 60))
		enMs := int(math.Mod(c.End*1000, 1000))

		sb.WriteString(fmt.Sprintf("%d\n", i+1))
		sb.WriteString(fmt.Sprintf("%02d:%02d:%02d,%03d --> %02d:%02d:%02d,%03d\n", stH, stM, stS, stMs, enH, enM, enS, enMs))
		sb.WriteString(c.Text + "\n\n")
	}
	return sb.String()
}
