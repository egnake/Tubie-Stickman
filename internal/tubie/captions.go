package tubie

import (
	"bytes"
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

// whisperScript reads its arguments as UTF-8 JSON from stdin (no string interpolation into code)
// and prints recognized words with acoustic timestamps as ASCII-safe JSON.
const whisperScript = `import json, sys
args = json.loads(sys.stdin.buffer.read().decode("utf-8"))
from faster_whisper import WhisperModel
model = WhisperModel(args["model"], device="cpu", compute_type="int8")
kwargs = dict(language="tr", word_timestamps=True, beam_size=5, condition_on_previous_text=False)
if args.get("prompt"):
    kwargs["initial_prompt"] = args["prompt"]
segments, _ = model.transcribe(args["wav"], **kwargs)
out = []
for s in segments:
    for w in (s.words or []):
        out.append({"word": w.word.strip(), "start": round(float(w.start), 3), "end": round(float(w.end), 3)})
sys.stdout.write(json.dumps(out, ensure_ascii=True))
`

// ExtractWhisperWords runs faster-whisper and returns recognized words with acoustic timings.
// The optional prompt (the known script) biases recognition towards the correct vocabulary.
func ExtractWhisperWords(wavPath string, prompt string) ([]TimedWord, error) {
	model := strings.TrimSpace(os.Getenv("TUBIE_WHISPER_MODEL"))
	if model == "" {
		model = "base"
	}
	if r := []rune(prompt); len(r) > 220 {
		prompt = string(r[:220])
	}
	payload, err := json.Marshal(map[string]string{"wav": wavPath, "prompt": prompt, "model": model})
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(BinaryPath("python"), "-c", whisperScript)
	cmd.Stdin = bytes.NewReader(payload)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var words []TimedWord
	if err := json.Unmarshal(bytes.TrimSpace(out), &words); err != nil {
		return nil, err
	}
	return words, nil
}

// ExtractWhisperTimestamps runs faster-whisper to extract acoustic timing intervals.
func ExtractWhisperTimestamps(wavPath string) ([]map[string]float64, error) {
	words, err := ExtractWhisperWords(wavPath, "")
	if err != nil {
		return nil, err
	}
	intervals := make([]map[string]float64, len(words))
	for i, w := range words {
		intervals[i] = map[string]float64{"start": w.Start, "end": w.End}
	}
	return intervals, nil
}

// SpeechSegment is a non-silent region of the narration audio.
type SpeechSegment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
}

var (
	silenceStartRe = regexp.MustCompile(`silence_start:\s*(-?[0-9.]+)`)
	silenceEndRe   = regexp.MustCompile(`silence_end:\s*(-?[0-9.]+)`)
)

// DetectSpeechSegments finds voiced regions of a WAV file using FFmpeg silencedetect.
// Returns nil when detection is unavailable.
func DetectSpeechSegments(wavPath string, totalDuration float64) []SpeechSegment {
	if wavPath == "" || totalDuration <= 0 {
		return nil
	}
	cmd := exec.Command(BinaryPath("ffmpeg"), "-hide_banner", "-nostats", "-i", wavPath,
		"-af", "silencedetect=noise=-38dB:d=0.18", "-f", "null", "-")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return nil
	}

	segments := make([]SpeechSegment, 0)
	cursor := 0.0
	inSilence := false
	for _, line := range strings.Split(string(out), "\n") {
		if m := silenceStartRe.FindStringSubmatch(line); m != nil {
			v := 0.0
			fmt.Sscanf(m[1], "%g", &v)
			v = math.Max(0, math.Min(v, totalDuration))
			if v-cursor >= 0.05 {
				segments = append(segments, SpeechSegment{Start: cursor, End: v})
			}
			inSilence = true
		} else if m := silenceEndRe.FindStringSubmatch(line); m != nil {
			v := 0.0
			fmt.Sscanf(m[1], "%g", &v)
			cursor = math.Max(cursor, math.Min(v, totalDuration))
			inSilence = false
		}
	}
	if !inSilence && totalDuration-cursor >= 0.05 {
		segments = append(segments, SpeechSegment{Start: cursor, End: totalDuration})
	}
	return segments
}

var turkishFold = strings.NewReplacer("ç", "c", "ğ", "g", "ı", "i", "ö", "o", "ş", "s", "ü", "u", "â", "a", "î", "i", "û", "u", "\u0307", "")

var (
	trOnes = []string{"sifir", "bir", "iki", "uc", "dort", "bes", "alti", "yedi", "sekiz", "dokuz"}
	trTens = []string{"", "on", "yirmi", "otuz", "kirk", "elli", "altmis", "yetmis", "seksen", "doksan"}
)

func turkishNumberKey(digits string) string {
	n := 0
	if len(digits) > 3 {
		return digits
	}
	for _, r := range digits {
		n = n*10 + int(r-'0')
	}
	if n == 0 {
		return trOnes[0]
	}
	var b strings.Builder
	if h := n / 100; h > 0 {
		if h > 1 {
			b.WriteString(trOnes[h])
		}
		b.WriteString("yuz")
	}
	b.WriteString(trTens[(n/10)%10])
	if o := n % 10; o > 0 {
		b.WriteString(trOnes[o])
	}
	return b.String()
}

// alignKey reduces a word to a diacritic-free lowercase key for fuzzy acoustic matching.
func alignKey(w string) string {
	lw := turkishFold.Replace(strings.ToLowerSpecial(unicode.TurkishCase, strings.TrimSpace(w)))
	var b strings.Builder
	allDigits := true
	for _, r := range lw {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
			if !unicode.IsDigit(r) {
				allDigits = false
			}
		}
	}
	k := b.String()
	if k != "" && allDigits {
		return turkishNumberKey(k)
	}
	return k
}

func levenshtein(a, b []rune) int {
	prev := make([]int, len(b)+1)
	cur := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		cur[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			cur[j] = min(prev[j]+1, cur[j-1]+1, prev[j-1]+cost)
		}
		prev, cur = cur, prev
	}
	return prev[len(b)]
}

// keySimilarity scores two alignment keys in [0,1], tolerant to Turkish suffix variations.
func keySimilarity(a, b string) float64 {
	if a == "" || b == "" {
		return 0
	}
	if a == b {
		return 1
	}
	ra, rb := []rune(a), []rune(b)
	short, long := ra, rb
	if len(short) > len(long) {
		short, long = long, short
	}
	if len(short) <= 2 {
		return 0 // very short words must match exactly
	}
	sim := 1 - float64(levenshtein(ra, rb))/float64(len(long))
	if len(short) >= 4 && strings.HasPrefix(string(long), string(short)) {
		sim = math.Max(sim, 0.75)
	}
	return sim
}

func wordWeight(w string) float64 {
	return float64(len([]rune(w))) + 2
}

// AlignScriptWords assigns acoustic timings to the canonical script words.
// It aligns recognized words to the script with dynamic programming (tolerating split, merged,
// misheard, missing and extra words), anchors matched words to their real timestamps, and
// interpolates the rest inside the surrounding anchors. Without recognized words it distributes
// words over detected speech regions (silences excluded).
func AlignScriptWords(canonical []string, recognized []TimedWord, speech []SpeechSegment, totalDuration float64) []TimedWord {
	n := len(canonical)
	if n == 0 {
		return []TimedWord{}
	}
	speechStart, speechEnd := 0.0, totalDuration
	if len(speech) > 0 {
		speechStart, speechEnd = speech[0].Start, speech[len(speech)-1].End
	}

	starts := make([]float64, n)
	ends := make([]float64, n)
	anchored := make([]bool, n)
	matched := 0

	if m := len(recognized); m > 0 {
		cKeys := make([]string, n)
		for i, w := range canonical {
			cKeys[i] = alignKey(w)
		}
		rKeys := make([]string, m)
		for j, w := range recognized {
			rKeys[j] = alignKey(w.Word)
		}

		type step struct{ di, dj int }
		const negInf = -1e18
		dp := make([][]float64, n+1)
		back := make([][]step, n+1)
		for i := range dp {
			dp[i] = make([]float64, m+1)
			back[i] = make([]step, m+1)
			for j := range dp[i] {
				dp[i][j] = negInf
			}
		}
		dp[0][0] = 0
		relax := func(i, j, di, dj int, score float64) {
			if prev := dp[i-di][j-dj]; prev > negInf && prev+score > dp[i][j]+1e-9 {
				dp[i][j] = prev + score
				back[i][j] = step{di, dj}
			}
		}
		for i := 0; i <= n; i++ {
			for j := 0; j <= m; j++ {
				if i == 0 && j == 0 {
					continue
				}
				if i > 0 {
					relax(i, j, 1, 0, 0)
				}
				if j > 0 {
					relax(i, j, 0, 1, 0)
				}
				if i == 0 || j == 0 {
					continue
				}
				// one script word heard as 1..3 recognized tokens
				joined := ""
				for k := 1; k <= 3 && j-k >= 0; k++ {
					joined = rKeys[j-k] + joined
					thr := 0.6
					if k > 1 {
						thr = 0.7
					}
					if s := keySimilarity(cKeys[i-1], joined); s >= thr {
						relax(i, j, 1, k, s-0.05*float64(k-1))
					}
				}
				// two script words heard as one recognized token
				if i >= 2 {
					if s := keySimilarity(cKeys[i-2]+cKeys[i-1], rKeys[j-1]); s >= 0.7 {
						relax(i, j, 2, 1, 2*s-0.05)
					}
				}
			}
		}

		for i, j := n, m; i > 0 || j > 0; {
			st := back[i][j]
			if st.di > 0 && st.dj > 0 {
				segStart := recognized[j-st.dj].Start
				segEnd := recognized[j-1].End
				if segEnd < segStart {
					segEnd = segStart
				}
				total := 0.0
				for c := i - st.di; c < i; c++ {
					total += wordWeight(canonical[c])
				}
				cursor := segStart
				for c := i - st.di; c < i; c++ {
					span := (segEnd - segStart) * wordWeight(canonical[c]) / total
					starts[c], ends[c], anchored[c] = cursor, cursor+span, true
					cursor += span
					matched++
				}
			}
			i -= st.di
			j -= st.dj
		}
	}

	timed := make([]TimedWord, n)
	if float64(matched) < math.Max(3, 0.35*float64(n)) {
		// Recognition unreliable or unavailable: distribute over voiced regions only.
		distributeOverSpeech(canonical, speech, speechStart, speechEnd, totalDuration, timed)
		return timed
	}

	// Average speaking pace from anchored words (seconds per weight unit).
	pace, paceWeight := 0.0, 0.0
	for i := range canonical {
		if anchored[i] {
			pace += ends[i] - starts[i]
			paceWeight += wordWeight(canonical[i])
		}
	}
	secPerWeight := 0.06
	if paceWeight > 0 && pace > 0 {
		secPerWeight = pace / paceWeight
	}

	for a := 0; a < n; {
		if anchored[a] {
			a++
			continue
		}
		b := a
		runWeight := 0.0
		for b < n && !anchored[b] {
			runWeight += wordWeight(canonical[b])
			b++
		}
		est := runWeight * secPerWeight * 1.3
		var left, right float64
		switch {
		case a > 0 && b < n:
			left, right = ends[a-1], starts[b]
		case a == 0 && b < n:
			right = starts[b]
			left = math.Max(math.Min(speechStart, right), right-est)
		case a > 0:
			left = ends[a-1]
			right = math.Min(math.Max(speechEnd, left), left+est)
		default:
			left, right = speechStart, speechEnd
		}
		if right-left < 0.06*float64(b-a) {
			right = math.Min(totalDuration, left+0.06*float64(b-a))
		}
		cursor := left
		for c := a; c < b; c++ {
			span := (right - left) * wordWeight(canonical[c]) / runWeight
			starts[c], ends[c] = cursor, cursor+span
			cursor += span
		}
		a = b
	}

	for i, w := range canonical {
		timed[i] = TimedWord{Word: w, Start: starts[i], End: ends[i]}
	}
	return timed
}

// distributeOverSpeech spreads words proportionally to their length across voiced audio time.
func distributeOverSpeech(canonical []string, speech []SpeechSegment, speechStart, speechEnd, totalDuration float64, timed []TimedWord) {
	segs := speech
	if len(segs) == 0 {
		lead := 0.2
		end := totalDuration - 0.4
		if end <= lead {
			lead, end = 0, totalDuration
		}
		segs = []SpeechSegment{{Start: lead, End: end}}
	}
	voiced := 0.0
	for _, s := range segs {
		voiced += s.End - s.Start
	}
	totalWeight := 0.0
	for _, w := range canonical {
		totalWeight += wordWeight(w)
	}
	// position maps a voiced-time offset to absolute audio time.
	position := func(offset float64, preferEarly bool) float64 {
		acc := 0.0
		for k, s := range segs {
			d := s.End - s.Start
			if offset < acc+d || (preferEarly && offset <= acc+d) || k == len(segs)-1 {
				return math.Min(s.End, s.Start+math.Max(0, offset-acc))
			}
			acc += d
		}
		return segs[len(segs)-1].End
	}
	cum := 0.0
	for i, w := range canonical {
		st := position(cum/totalWeight*voiced, false)
		cum += wordWeight(w)
		en := position(cum/totalWeight*voiced, true)
		if en < st {
			en = st
		}
		timed[i] = TimedWord{Word: w, Start: st, End: en}
	}
}

// BuildAccurateCaptions creates perfectly spelled, acoustically aligned viral captions from the script.
func BuildAccurateCaptions(scriptText string, wavPath string, totalDuration float64) []CaptionCue {
	canonicalWords := ExtractScriptWords(scriptText)
	if len(canonicalWords) == 0 {
		return []CaptionCue{}
	}

	var recognized []TimedWord
	var speech []SpeechSegment
	if wavPath != "" {
		if fi, err := os.Stat(wavPath); err == nil && !fi.IsDir() && fi.Size() > 0 {
			if words, err := ExtractWhisperWords(wavPath, scriptText); err == nil {
				recognized = words
			}
			speech = DetectSpeechSegments(wavPath, totalDuration)
		}
	}

	return GroupCaptionCues(AlignScriptWords(canonicalWords, recognized, speech, totalDuration), totalDuration)
}

// GroupCaptionCues groups timed words into monotonic 1-3 word viral caption cues.
func GroupCaptionCues(timedWords []TimedWord, totalDuration float64) []CaptionCue {
	for k := range timedWords {
		timedWords[k].Start = math.Round(timedWords[k].Start*1000) / 1000
		timedWords[k].End = math.Round(timedWords[k].End*1000) / 1000
	}

	// 3. Group into 1-3 word cues (Alex Hormozi viral rhythm), never across a real pause
	const pauseBreak = 0.35
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
		for k := 1; k < take; k++ {
			if timedWords[i+k].Start-timedWords[i+k-1].End > pauseBreak {
				take = k
				break
			}
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
	// Close tiny gaps between consecutive cues so captions don't flicker between words.
	for k := 0; k+1 < len(rawCues); k++ {
		if gap := rawCues[k+1].Start - rawCues[k].End; gap > 0 && gap <= pauseBreak {
			rawCues[k].End = rawCues[k+1].Start
		}
	}

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

		// Keep real per-word timings (for karaoke highlight), clamped inside the cue and monotonic
		words := make([]TimedWord, len(cue.Words))
		prevEnd := st
		for wi, w := range cue.Words {
			wSt := math.Min(math.Max(w.Start, prevEnd), en)
			wEn := math.Min(math.Max(w.End, wSt), en)
			if wi == len(cue.Words)-1 {
				wEn = en
			}
			words[wi] = TimedWord{
				Word:  w.Word,
				Start: math.Round(wSt*1000) / 1000,
				End:   math.Round(wEn*1000) / 1000,
			}
			prevEnd = wEn
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
