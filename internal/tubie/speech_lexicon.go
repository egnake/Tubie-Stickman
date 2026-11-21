package tubie

import (
	"strings"
	"unicode"
)

var multiWordTechPronunciations = []struct {
	w1, w2 string
	spoken string
}{
	{"prepared", "statements", "pripeyrd steytmınts"},
	{"prepared", "statement", "pripeyrd steytmınt"},
	{"prompt", "injection", "promt incekşın"},
	{"sql", "injection", "es kü el incekşın"},
	{"brute", "force", "brut fors"},
	{"evil", "twin", "ivil tvin"},
	{"sim", "swap", "sim svap"},
	{"zero", "day", "ziro dey"},
	{"dark", "web", "dark veb"},
	{"seed", "phrase", "sid freyz"},
	{"two", "factor", "tu fektır"},
}

// strictTechPronunciations matches exact uppercase acronyms or canonical mixed-case tech names
// so plain Turkish lowercase words (like "ip" = rope) are never altered.
var strictTechPronunciations = map[string]string{
	"APK":   "a pe ka",
	"2FA":   "iki ef ey",
	"MFA":   "em ef ey",
	"QR":    "kü ar",
	"URL":   "u re le",
	"VPN":   "vi pi en",
	"SQL":   "es kü el",
	"SQLI":  "es kü el i",
	"SQLi":  "es kü el i",
	"IOT":   "ay o ti",
	"IoT":   "ay o ti",
	"IOS":   "ay o es",
	"iOS":   "ay o es",
	"AI":    "ey ay",
	"DM":    "di em",
	"OTP":   "o ti pi",
	"SMS":   "es em es",
	"IP":    "ay pi",
	"DNS":   "di en es",
	"SSH":   "es es eyç",
	"HTTP":  "eyç ti ti pi",
	"HTTPS": "eyç ti ti pi es",
	"SSL":   "es es el",
	"TLS":   "ti el es",
	"API":   "ey pi ay",
	"PDF":   "pi di ef",
	"USB":   "yu es bi",
	"LLM":   "el el em",
	"XSS":   "iks es es",
	"DDOS":  "di dos",
	"DDoS":  "di dos",
	"MITM":  "em ay ti em",
	"UPNP":  "yu pi en pi",
	"UPnP":  "yu pi en pi",
	"NFC":   "en ef si",
	"GPS":   "ci pi es",
	"CVE":   "si vi i",
	"PMKID": "pi em key ay di",
	"WPA":   "dabılyu pi ey",
	"Wi-Fi": "vayfay",
	"WiFi":  "vayfay",
	"WIFI":  "vayfay",
}

// techLoanwordPronunciations matches unambiguous English/cyber loanwords case-insensitively.
var techLoanwordPronunciations = map[string]string{
	"crack":         "kırek",
	"phishing":      "fişing",
	"quishing":      "küişing",
	"deepfake":      "dipfeyk",
	"infostealer":   "infostilır",
	"malware":       "malveyr",
	"ransomware":    "rensımveyr",
	"spyware":       "spayveyr",
	"firmware":      "förmveyr",
	"firewall":      "fayırvol",
	"passkey":       "peski",
	"authenticator": "otentikeytır",
	"deauth":        "diot",
	"airdrop":       "eyırdrop",
	"approve":       "epruv",
	"drainer":       "dreynır",
	"cookie":        "kuki",
	"bluetooth":     "blutut",
	"whatsapp":      "vatsap",
	"youtube":       "yutub",
	"google":        "gugıl",
	"chrome":        "krom",
	"apple":         "epıl",
	"iphone":        "ayfon",
	"chatgpt":       "çet ci pi ti",
	"hacker":        "hekır",
	"hackle":        "hekle",
	"exploit":       "eksployt",
	"payload":       "peylod",
	"keylogger":     "ki logır",
	"backdoor":      "bekdor",
	"router":        "rutır",
	"hotspot":       "hatspot",
	"token":         "tokın",
	"hash":          "heş",
	"session":       "seşın",
	"cloud":         "klaud",
	"backup":        "bekap",
	"scam":          "skem",
	"fake":          "feyk",
	"web3":          "veb üç",
	"wallet":        "volıt",
	"update":        "apdeyt",
	"patch":         "peç",
	"admin":         "edmin",
	"string":        "sıtring",
	"wifi":          "vayfay",
	"wi-fi":         "vayfay",
	"zero-day":      "ziro dey",
}

// agglutinableStemsOrdered lists unambiguous stems (len >= 5) sorted longest-first
// for agglutinated Turkish inflections written without an apostrophe (e.g. "crackli", "hackerlar").
var agglutinableStemsOrdered = []struct {
	stem   string
	spoken string
}{
	{"authenticator", "otentikeytır"},
	{"infostealer", "infostilır"},
	{"ransomware", "rensımveyr"},
	{"keylogger", "ki logır"},
	{"deepfake", "dipfeyk"},
	{"phishing", "fişing"},
	{"quishing", "küişing"},
	{"firmware", "förmveyr"},
	{"firewall", "fayırvol"},
	{"backdoor", "bekdor"},
	{"whatsapp", "vatsap"},
	{"malware", "malveyr"},
	{"spyware", "spayveyr"},
	{"passkey", "peski"},
	{"airdrop", "eyırdrop"},
	{"approve", "epruv"},
	{"drainer", "dreynır"},
	{"exploit", "eksployt"},
	{"payload", "peylod"},
	{"session", "seşın"},
	{"youtube", "yutub"},
	{"cookie", "kuki"},
	{"hacker", "hekır"},
	{"hackle", "hekle"},
	{"router", "rutır"},
	{"iphone", "ayfon"},
	{"backup", "bekap"},
	{"deauth", "diot"},
	{"chrome", "krom"},
	{"google", "gugıl"},
	{"wallet", "volıt"},
	{"update", "apdeyt"},
	{"crack", "kırek"},
	{"token", "tokın"},
	{"apple", "epıl"},
	{"admin", "edmin"},
}

func splitEdgePunctuation(tok string) (lead, core, trail string) {
	runes := []rune(tok)
	start := 0
	for start < len(runes) && !unicode.IsLetter(runes[start]) && !unicode.IsDigit(runes[start]) {
		start++
	}
	end := len(runes)
	for end > start && !unicode.IsLetter(runes[end-1]) && !unicode.IsDigit(runes[end-1]) {
		end--
	}
	return string(runes[:start]), string(runes[start:end]), string(runes[end:])
}

func cutApostrophe(s string) (stem, sep, suffix string, ok bool) {
	for _, candidate := range []string{"'", "’"} {
		if before, after, found := strings.Cut(s, candidate); found && before != "" {
			return before, candidate, after, true
		}
	}
	return s, "", "", false
}

func lookupStrictOrLoanword(stem string) (string, bool) {
	if spoken, ok := strictTechPronunciations[stem]; ok {
		return spoken, true
	}
	lower := strings.ToLowerSpecial(unicode.TurkishCase, stem)
	if spoken, ok := techLoanwordPronunciations[lower]; ok {
		return spoken, true
	}
	// Check agglutinated stems written without apostrophe (e.g. crackli -> kırekli, hackerlar -> hekırlar)
	for _, entry := range agglutinableStemsOrdered {
		if len(lower) > len(entry.stem) && strings.HasPrefix(lower, entry.stem) {
			rest := lower[len(entry.stem):]
			if isTurkishSuffixLike(rest) {
				return entry.spoken + rest, true
			}
		}
	}
	return "", false
}

func isTurkishSuffixLike(s string) bool {
	if s == "" {
		return true
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

func rewriteSpeechToken(tok string) string {
	lead, core, trail := splitEdgePunctuation(tok)
	if core == "" {
		return tok
	}
	stem, sep, suffix, hasApos := cutApostrophe(core)
	if spoken, ok := lookupStrictOrLoanword(stem); ok {
		if hasApos && suffix != "" {
			return lead + spoken + sep + suffix + trail
		}
		return lead + spoken + trail
	}
	return tok
}

// PrepareSpeechForTurkishTTS converts cybersecurity acronyms and English tech terms
// into natural Turkish phonetic spellings for Edge-TTS / XTTS synthesis while preserving
// surrounding punctuation, Turkish suffixes, and normal Turkish words.
func PrepareSpeechForTurkishTTS(text string) string {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return ""
	}
	out := make([]string, 0, len(fields))
	for i := 0; i < len(fields); {
		if i+1 < len(fields) {
			l1, c1, t1 := splitEdgePunctuation(fields[i])
			l2, c2, t2 := splitEdgePunctuation(fields[i+1])
			// Only merge across space when the first token does not end with sentence/clause punctuation.
			if c1 != "" && t1 == "" && c2 != "" && l2 == "" && !strings.ContainsAny(c1, "'’") {
				s1 := strings.ToLowerSpecial(unicode.TurkishCase, c1)
				stem2, sep2, suf2, hasApos2 := cutApostrophe(c2)
				s2 := strings.ToLowerSpecial(unicode.TurkishCase, stem2)
				matched := ""
				for _, p := range multiWordTechPronunciations {
					if s1 == p.w1 && s2 == p.w2 {
						matched = p.spoken
						break
					}
				}
				if matched != "" {
					if hasApos2 && suf2 != "" {
						out = append(out, l1+matched+sep2+suf2+t2)
					} else {
						out = append(out, l1+matched+t2)
					}
					i += 2
					continue
				}
			}
		}
		out = append(out, rewriteSpeechToken(fields[i]))
		i++
	}
	return strings.Join(out, " ")
}

// PhoneticAlignKey converts a canonical uppercase caption token (e.g. "APK", "CRACK", "APKYI", "WI-FI")
// into a normalized alignment key matching how Turkish TTS speaks it so Whisper acoustic alignment
// anchors acronyms and loanwords reliably.
func PhoneticAlignKey(canonicalWord string) string {
	trimmed := strings.TrimSpace(canonicalWord)
	if trimmed == "" {
		return ""
	}
	if spoken := PrepareSpeechForTurkishTTS(trimmed); spoken != trimmed {
		return alignKey(spoken)
	}
	// Handle stripped apostrophe suffixes (e.g. "APKYI" -> "APK" + "YI", "CRACKI" -> "CRACK" + "I")
	if spoken, ok := strictTechPronunciations[trimmed]; ok {
		return alignKey(spoken)
	}
	for k, v := range strictTechPronunciations {
		if len(k) >= 3 && strings.HasPrefix(trimmed, k) && len(trimmed) > len(k) {
			rest := trimmed[len(k):]
			if isTurkishSuffixLike(rest) {
				return alignKey(v + rest)
			}
		}
	}
	return alignKey(trimmed)
}
