package tubie

import (
	"math"
	"strings"
	"unicode"
)

// ViralDimension scores a single aspect of short-form cybersecurity video retention.
type ViralDimension struct {
	Name     string `json:"name"`
	Score    int    `json:"score"`
	MaxScore int    `json:"max_score"`
	Feedback string `json:"feedback"`
}

// ViralScoreReport provides an actionable retention and viral readiness breakdown.
type ViralScoreReport struct {
	TotalScore       int              `json:"total_score"`
	Grade            string           `json:"grade"`
	EstimatedSeconds int              `json:"estimated_seconds"`
	WordCount        int              `json:"word_count"`
	HookSentence     string           `json:"hook_sentence"`
	ThreatCategory   string           `json:"threat_category"`
	LoopReady        bool             `json:"loop_ready"`
	Dimensions       []ViralDimension `json:"dimensions"`
	Tips             []string         `json:"tips"`
}

// CyberIdea represents a trending, high-CTR cybersecurity video concept for the Studio deck.
type CyberIdea struct {
	ID               string `json:"id"`
	Category         string `json:"category"`
	Title            string `json:"title"`
	Prompt           string `json:"prompt"`
	HookPreview      string `json:"hook_preview"`
	RecommendedStyle string `json:"recommended_style"`
}

// DetectThreatCategory classifies a title/narration into a Turkish cybersecurity domain.
func DetectThreatCategory(text string) string {
	norm := normalizeMatchText(text)
	switch {
	case containsAnyKeyword(norm, []string{"yapay zeka", "chatgpt", "deepfake", "ses klon", "prompt", "llm", "ajan", "model"}):
		return "Yapay Zeka & Deepfake"
	case containsAnyKeyword(norm, []string{"wifi", "wi-fi", "evil twin", "deauth", "vpn", "modem", "router", "bluetooth", "airtag", "iot", "akilli ev", "akilli cihaz", "kamera", "termostat", "priz", "botnet", "mirai"}):
		return "Ağ, Wi-Fi & Kablosuz"
	case containsAnyKeyword(norm, []string{"kripto", "bitcoin", "cuzdan", "airdrop", "metamask", "seed", "token", "nft"}):
		return "Kripto & Web3 Güvenliği"
	case containsAnyKeyword(norm, []string{"crack", "apk", "virus", "trojan", "fidye", "ransomware", "infostealer", "eklenti", "usb"}):
		return "Zararlı Yazılım & Cihaz"
	case containsAnyKeyword(norm, []string{"sql", "injection", "xss", "api", "veritabani", "sunucu", "kod", "dark web", "darkweb"}):
		return "Web, Veri & Kod Güvenliği"
	case containsAnyKeyword(norm, []string{"oltalama", "phishing", "sahte link", "instagram", "whatsapp", "sms", "sim", "e-devlet", "banka", "qr", "quishing"}):
		return "Oltalama & Sosyal Mühendislik"
	default:
		return "Dijital Gizlilik & Savunma"
	}
}

func firstSentenceOf(text string) string {
	cleaned := cleanMarkdownInline(text)
	if cleaned == "" {
		return ""
	}
	for i, r := range cleaned {
		if r == '.' || r == '!' || r == '?' || r == '…' {
			part := strings.TrimSpace(cleaned[:i+1])
			if len([]rune(part)) >= 4 {
				return part
			}
		}
	}
	return cleaned
}

func lastSentenceOf(text string) string {
	sents := splitSentences(text)
	if len(sents) == 0 {
		return strings.TrimSpace(text)
	}
	return sents[len(sents)-1]
}

// AnalyzeScriptVirality evaluates hook strength, tension arc, pacing, specificity, CTA, and loop flow.
func AnalyzeScriptVirality(title, narration string, scenes []SceneDefinition, loopEnabled bool) ViralScoreReport {
	cleanNarr := cleanMarkdownInline(narration)
	words := strings.Fields(cleanNarr)
	wordCount := len(words)
	estSeconds := int(math.Round(float64(wordCount) / 2.65))
	if wordCount > 0 && estSeconds < 5 {
		estSeconds = 5
	}

	hook := firstSentenceOf(cleanNarr)
	normAll := normalizeMatchText(title + " " + cleanNarr)
	normHook := normalizeMatchText(hook + " " + title)
	normTail := normalizeMatchText(lastSentenceOf(cleanNarr))
	tips := make([]string, 0, 5)

	// 1. Hook Strength (0-25)
	hookScore := 6
	hookWords := len(strings.Fields(hook))
	if hookWords >= 4 && hookWords <= 14 {
		hookScore += 6
	} else if hookWords > 18 {
		tips = append(tips, "İlk cümleni 10-12 kelimenin altına kısalt; izleyici ilk 3 saniyede yakalanmalı.")
	} else if hookWords > 0 {
		hookScore += 3
	}
	if strings.Contains(hook, "?") || strings.Contains(hook, "!") || strings.Contains(hook, "…") {
		hookScore += 4
	}
	if containsAnyKeyword(normHook, []string{"sen", "seni", "senin", "telefon", "ekran", "hesab", "cuzdan", "saniye", "dikkat", "asla", "nasil", "gercek", "tuzak", "tehlike", "yanlis", "tek bir", "bir bildirim"}) {
		hookScore += 7
	} else {
		tips = append(tips, "Açılış cümlesine doğrudan izleyiciyi merkeze alan ('senin hesabın', '3 saniyede...') bir kanca ekle.")
	}
	if containsAnyKeyword(normHook, []string{"merhaba", "bugun sizlerle", "hos geldiniz", "videoma hos", "kanalima", "selam arkadaslar"}) {
		hookScore -= 10
		tips = append(tips, "'Merhaba / Bugün sizlerle' gibi yavaş girişleri kaldırıp doğrudan kriz anıyla başla.")
	}
	if containsAnyKeyword(normHook, []string{"sandigindan cok daha kritik"}) {
		hookScore -= 5
	}
	hookScore = clampInt(hookScore, 0, 25)

	// 2. Tension & Story Arc (0-20)
	arcScore := 4
	hasCrisis := containsAnyKeyword(normAll, []string{"calin", "ele gec", "tuzak", "tehlike", "sizinti", "saldiri", "kilit", "bosal", "sifirlan", "dinlen", "izlen"})
	hasPivot := containsAnyKeyword(normAll, []string{"oysa", "peki ya", "ama", "cunku", "gercek su", "gorunuyor", "fark etmeden", "arka planda"})
	hasDefense := containsAnyKeyword(normAll, []string{"koru", "kapat", "kullan", "dogrula", "onlem", "kural", "aktif", "kontrol et", "sadece"})
	if hasCrisis {
		arcScore += 6
	} else {
		tips = append(tips, "Somut risk veya kayıp senaryosunu netleştir (ör. oturum çalınması, veri sızıntısı).")
	}
	if hasPivot {
		arcScore += 5
	} else {
		tips = append(tips, "'Oysa...', 'Peki ya...' gibi bir merak/kırılma bağlacı ile gerilimi tırmandır.")
	}
	if hasDefense {
		arcScore += 5
	} else {
		tips = append(tips, "Videonun sonunda izleyicinin hemen uygulayabileceği net bir savunma çözümü ver.")
	}
	arcScore = clampInt(arcScore, 0, 20)

	// 3. Pacing & Rhythm (0-15)
	pacingScore := 0
	switch {
	case wordCount >= 90 && wordCount <= 165:
		pacingScore += 9
	case wordCount >= 65 && wordCount < 90:
		pacingScore += 6
		tips = append(tips, "Senaryo biraz kısa; 105-145 kelime aralığı Shorts/Reels tutma süresini optimize eder.")
	case wordCount > 165 && wordCount <= 200:
		pacingScore += 6
		tips = append(tips, "Metin 60 saniyeyi zorlayabilir; gereksiz dolgu kelimelerini buda.")
	case wordCount >= 35:
		pacingScore += 3
		tips = append(tips, "İdeal kısa video temposu için metni 100-145 kelime aralığına getir.")
	default:
		tips = append(tips, "Seslendirme metni çok kısa; en az 90-140 kelimelik akıcı bir anlatım yaz.")
	}
	sents := splitSentences(cleanNarr)
	avgSentWords := float64(wordCount)
	if len(sents) > 0 {
		avgSentWords = float64(wordCount) / float64(len(sents))
	}
	if len(sents) >= 5 && avgSentWords >= 6 && avgSentWords <= 16 {
		pacingScore += 6
	} else if len(sents) >= 3 && avgSentWords <= 22 {
		pacingScore += 3
	} else {
		tips = append(tips, "Cümleleri daha kısa ve vurucu tut (ortalama 8-13 kelime nefes temposunu hızlandırır).")
	}
	pacingScore = clampInt(pacingScore, 0, 15)

	// 4. Pattern Interrupts & Technical Concrete Details (0-15)
	techScore := 0
	techTokens := []string{
		"2fa", "iki asamali", "iki adimli", "authenticator", "vpn", "https", "otp", "passkey", "mfa",
		"sql", "evil twin", "deauth", "infostealer", "ransomware", "fidye", "phishing", "oltalama",
		"quishing", "qr", "deepfake", "sim", "cerez", "cookie", "virustotal", "brute force",
		"kaba kuvvet", "zero-day", "prompt injection", "api", "usb", "approve", "sifre yoneticisi",
		"parola yoneticisi", "soguk cuzdan", "cold wallet", "parametrik", "pin",
		"iot", "firmware", "upnp", "vlan", "misafir agi", "varsayilan sifre", "botnet", "mirai", "ddos", "zero trust",
	}
	techHits := scoreKeywords(normAll, techTokens)
	if techHits >= 3 {
		techScore += 6
	} else if techHits >= 1 {
		techScore += 3
	} else {
		tips = append(tips, "Genel öğütler yerine somut siber teknikler (örn. 2FA, Evil Twin, Passkey, OTP, Infostealer) ekle.")
	}

	hasNumbers := containsAnyKeyword(normAll, []string{"bir", "iki", "uc", "dort", "bes", "alti", "saniye", "dakika", "14", "16"})
	if !hasNumbers {
		for _, r := range cleanNarr {
			if unicode.IsDigit(r) {
				hasNumbers = true
				break
			}
		}
	}
	if hasNumbers {
		techScore += 4
	}
	if len(scenes) >= 5 {
		uniqueVisuals := make(map[string]bool)
		for _, sc := range scenes {
			vt := strings.TrimSpace(sc.VisualType)
			if vt != "" {
				uniqueVisuals[vt] = true
			}
		}
		if len(uniqueVisuals) >= 3 || len(scenes) >= 6 {
			techScore += 5
		} else {
			techScore += 3
		}
	}
	techScore = clampInt(techScore, 0, 15)

	// 5. Saveability & Engagement CTA (0-15)
	ctaScore := 3
	if len(extractNumberedRules(cleanNarr)) >= 2 || containsAnyKeyword(normAll, []string{"birinci kural", "ikinci kural", "uc altin", "uc kural", "dort kontrol", "uc temel"}) {
		ctaScore += 7
	} else {
		tips = append(tips, "'Bir: ..., İki: ..., Üç: ...' şeklinde numaralı adımlar izleyicinin videoyu kaydetme oranını 2-3x artırır.")
	}
	if containsAnyKeyword(normAll, []string{"kaydet", "paylas", "gonder", "yorum", "uyar", "etiketle", "takipte kal", "tekrar sor", "basa don"}) {
		ctaScore += 5
	} else {
		tips = append(tips, "Sona doğal bir kaydetme/paylaşma veya yorum çağrısı ekle.")
	}
	ctaScore = clampInt(ctaScore, 0, 15)

	// 6. Seamless Loop Readiness (0-10)
	loopScore := 0
	isLoopFormatted := containsAnyKeyword(normTail, []string{
		"iste bu yuzden", "tam da bu yuzden", "bir dahaki sefere", "simdi tekrar",
		"basa don", "unutma cunku", "soruyu tekrar", "cevabi su",
	}) || strings.HasSuffix(strings.TrimSpace(cleanNarr), "…")
	hasHardSignoff := containsAnyKeyword(normTail, []string{"takipte kal", "gorusuruz", "hosca kal", "abone ol"})

	if loopEnabled {
		if isLoopFormatted && !hasHardSignoff {
			loopScore = 10
		} else if !hasHardSignoff {
			loopScore = 6
			tips = append(tips, "Sonsuz döngü için son cümleyi ilk cümleye bağla (örn. 'İşte bu yüzden şimdi kendine şunu sor:').")
		} else {
			loopScore = 2
			tips = append(tips, "Sonsuz döngü modunda 'Takipte kal' gibi vedalar döngüyü bozar; son cümleyi giriş kancasına bağla.")
		}
	} else {
		if isLoopFormatted || (hasHardSignoff && hasDefense) {
			loopScore = 9
		} else if hasDefense {
			loopScore = 7
		} else {
			loopScore = 4
		}
	}
	loopScore = clampInt(loopScore, 0, 10)

	total := clampInt(hookScore+arcScore+pacingScore+techScore+ctaScore+loopScore, 0, 100)
	grade := "D"
	switch {
	case total >= 90:
		grade = "S"
	case total >= 78:
		grade = "A"
	case total >= 62:
		grade = "B"
	case total >= 45:
		grade = "C"
	}

	if len(tips) > 4 {
		tips = tips[:4]
	}

	return ViralScoreReport{
		TotalScore:       total,
		Grade:            grade,
		EstimatedSeconds: estSeconds,
		WordCount:        wordCount,
		HookSentence:     hook,
		ThreatCategory:   DetectThreatCategory(title + " " + cleanNarr),
		LoopReady:        isLoopFormatted && !hasHardSignoff,
		Dimensions: []ViralDimension{
			{Name: "Kanca Gücü (0-3 sn)", Score: hookScore, MaxScore: 25, Feedback: dimFeedback(hookScore, 25)},
			{Name: "Gerilim & Hikâye Yayı", Score: arcScore, MaxScore: 20, Feedback: dimFeedback(arcScore, 20)},
			{Name: "Tempo & Ritim", Score: pacingScore, MaxScore: 15, Feedback: dimFeedback(pacingScore, 15)},
			{Name: "Teknik Somutluk & Kırıcı", Score: techScore, MaxScore: 15, Feedback: dimFeedback(techScore, 15)},
			{Name: "Kaydetme & Etkileşim", Score: ctaScore, MaxScore: 15, Feedback: dimFeedback(ctaScore, 15)},
			{Name: "Kapanış / Döngü Uyumu", Score: loopScore, MaxScore: 10, Feedback: dimFeedback(loopScore, 10)},
		},
		Tips: tips,
	}
}

func dimFeedback(score, max int) string {
	ratio := float64(score) / float64(max)
	switch {
	case ratio >= 0.85:
		return "Çok Güçlü"
	case ratio >= 0.65:
		return "İyi"
	case ratio >= 0.45:
		return "Geliştirilebilir"
	default:
		return "Zayıf"
	}
}

func clampInt(v, minV, maxV int) int {
	if v < minV {
		return minV
	}
	if v > maxV {
		return maxV
	}
	return v
}

// TrendingCyberIdeas returns a curated deck of high-CTR cybersecurity video prompts grouped by domain.
func TrendingCyberIdeas() []CyberIdea {
	return []CyberIdea{
		{
			ID:               "idea-phishing-5sec",
			Category:         "Oltalama & Sosyal Mühendislik",
			Title:            "Bu Linke Tıklarsan Ne Olur?",
			Prompt:           "Sahte linkleri 5 saniyede nasıl anlarsın? Kilit simgesi ve küçük harf tuzakları",
			HookPreview:      "Kilit simgesi gördüğün site aslında tamamen sahte olabilir!",
			RecommendedStyle: StyleHybrid,
		},
		{
			ID:               "idea-edevlet-sms",
			Category:         "Oltalama & Sosyal Mühendislik",
			Title:            "Sahte E-Devlet ve Banka SMS Tuzağı",
			Prompt:           "Telefonuna gelen sahte e-devlet vergi iadesi ve kargo borcu SMS tuzağı nasıl çalışır?",
			HookPreview:      "Adın soyadın doğru yazıyor diye o SMS'i gerçek sanma!",
			RecommendedStyle: StyleCinematic,
		},
		{
			ID:               "idea-instagram-dm",
			Category:         "Oltalama & Sosyal Mühendislik",
			Title:            "Tek DM İle Instagram Çalma Tuzağı",
			Prompt:           "İnstagram telif hakkı ihlali ve mavi tik DM hesap çalma tuzağı",
			HookPreview:      "24 saat içinde hesabın kapatılacak diyen o DM tek tuşla her şeyi alıyor.",
			RecommendedStyle: StyleHybrid,
		},
		{
			ID:               "idea-qr-quishing",
			Category:         "Oltalama & Sosyal Mühendislik",
			Title:            "Sokaktaki QR Kodun Altındaki Tuzak",
			Prompt:           "Otopark ve restoranlardaki sahte QR kod (Quishing) ödeme dolandırıcılığı",
			HookPreview:      "Gözün iki karekod arasındaki farkı asla göremez!",
			RecommendedStyle: StyleMythBuster,
		},
		{
			ID:               "idea-deepfake-voice",
			Category:         "Yapay Zeka & Deepfake",
			Title:            "3 Saniyelik Sesinle Ailen Nasıl Kandırılır?",
			Prompt:           "Yapay zeka ile 3 saniyelik videodan ses klonlama (deepfake) telefon dolandırıcılığı",
			HookPreview:      "Gece yarısı ağlayarak para isteyen ses gerçekten kardeşin olmayabilir.",
			RecommendedStyle: StyleCinematic,
		},
		{
			ID:               "idea-ai-privacy",
			Category:         "Yapay Zeka & Deepfake",
			Title:            "Yapay Zeka Seni İnternette Ne Kadar Tanıyor?",
			Prompt:           "Yapay zeka sohbet geçmişinle dijital profil çıkarıyor mu? Gizlilik ayarları",
			HookPreview:      "Sohbet kutusuna yazdığın her cümle kalıcı bir profile dönüşüyor.",
			RecommendedStyle: StyleInformative,
		},
		{
			ID:               "idea-prompt-injection",
			Category:         "Yapay Zeka & Deepfake",
			Title:            "Görünmez Yazıyla Yapay Zekayı Hacklemek",
			Prompt:           "LLM Prompt Injection saldırısı nedir? Yapay zeka ajanları beyaz yazıyla nasıl kandırılır?",
			HookPreview:      "Ekranda görmediğin tek satır gizli yazı yapay zeka asistanını ele geçiriyor!",
			RecommendedStyle: StyleRedVsBlue,
		},
		{
			ID:               "idea-wifi-eviltwin",
			Category:         "Ağ, Wi-Fi & Kablosuz",
			Title:            "Kafe Wi-Fi Ağında Yan Masadaki Hacker",
			Prompt:           "Kafe ücretsiz Wi-Fi ağlarında Deauth ve Evil Twin sahte ikiz ağ saldırısı",
			HookPreview:      "Aynı Wi-Fi adını açıp tüm trafiğini iki masa öteden izliyorlar!",
			RecommendedStyle: StyleRedVsBlue,
		},
		{
			ID:               "idea-bluetooth-stalk",
			Category:         "Ağ, Wi-Fi & Kablosuz",
			Title:            "Bluetooth Açık Gezerken Takip Ediliyor musun?",
			Prompt:           "Bluetooth ve çanta içi takip cihazlarıyla (AirTag) konum izleme riskleri ve korunma",
			HookPreview:      "Kulaklık bağlamak için açık unuttuğun o ayar adım adım seni ele verebilir.",
			RecommendedStyle: StyleChallenge,
		},
		{
			ID:               "idea-iot-camera",
			Category:         "Ağ, Wi-Fi & Kablosuz",
			Title:            "Evindeki Güvenlik Kamerası İnternete Açık Olabilir",
			Prompt:           "Akıllı ev güvenlik kamerası ve modem varsayılan şifre açıkları nasıl bulunur?",
			HookPreview:      "Kutudan çıktığı şifreyle bıraktığın kamera herkese açık yayında olabilir!",
			RecommendedStyle: StyleMythBuster,
		},
		{
			ID:               "idea-infostealer-crack",
			Category:         "Zararlı Yazılım & Cihaz",
			Title:            "Antivirüsü Kapat Diyen Crack Dosyası",
			Prompt:           "Ücretsiz crack ve modlu APK dosyalarındaki sessiz Infostealer şifre hırsızlığı",
			HookPreview:      "Bilgisayarın yavaşlamaz bile ama tarayıcıdaki tüm şifrelerin 3 saniyede gider!",
			RecommendedStyle: StyleCinematic,
		},
		{
			ID:               "idea-ransomware",
			Category:         "Zararlı Yazılım & Cihaz",
			Title:            "Tüm Dosyaların Bir Gecede Nasıl Kilitlenir?",
			Prompt:           "Fidye yazılımı (Ransomware) saldırısı nasıl bulaşır ve 3-2-1 yedekleme nasıl kurtarır?",
			HookPreview:      "Tek bir e-posta ekiyle yılların fotoğrafları kilitlenip fidye istenebilir!",
			RecommendedStyle: StyleHybrid,
		},
		{
			ID:               "idea-usb-juice",
			Category:         "Zararlı Yazılım & Cihaz",
			Title:            "Havalimanındaki Ücretsiz USB Şarj Tuzağı",
			Prompt:           "Halka açık USB şarj noktalarında Juice Jacking veri hırsızlığı ve kötü amaçlı kablo",
			HookPreview:      "Telefonunu şarj ettiğini sanırken aslında veri hattını açıyor olabilirsin!",
			RecommendedStyle: StyleMythBuster,
		},
		{
			ID:               "idea-browser-ext",
			Category:         "Zararlı Yazılım & Cihaz",
			Title:            "Masum Tarayıcı Eklentisi Çerezlerini Çalıyor!",
			Prompt:           "Zararlı Chrome tarayıcı eklentileriyle oturum çerezi (Session Cookie) hırsızlığı",
			HookPreview:      "Şifreni ve 2FA kodunu bilmeden bile hesabına giriş yapabilirler!",
			RecommendedStyle: StyleChallenge,
		},
		{
			ID:               "idea-sim-swap",
			Category:         "Mobil & Kripto",
			Title:            "Telefonun Çekmiyorsa Hesapların Gidiyor Olabilir!",
			Prompt:           "SIM Swapping (SIM kart kopyalama) ile SMS 2FA aşma ve banka hesabı boşaltma",
			HookPreview:      "Telefonun bir anda 'Servis Yok' yazarsa sebebi arıza değil SIM saldırısı olabilir!",
			RecommendedStyle: StyleCinematic,
		},
		{
			ID:               "idea-crypto-drainer",
			Category:         "Mobil & Kripto",
			Title:            "Tek Onayla Boşaltılan Kripto Cüzdanları",
			Prompt:           "Sahte airdrop ve sınırsız Approve imzasıyla kripto cüzdan boşaltma tuzağı",
			HookPreview:      "12 gizli kelimeni çalmadan tek bir onay imzasıyla tüm bakiyeyi çekiyorlar!",
			RecommendedStyle: StyleRedVsBlue,
		},
		{
			ID:               "idea-sqli",
			Category:         "Web, Veri & Kod Güvenliği",
			Title:            "Tek Tırnak İşaretiyle Veritabanı Düşürmek",
			Prompt:           "SQL Injection açığı nasıl çalışır ve parametrik sorgularla nasıl engellenir?",
			HookPreview:      "Şifre bilmeden sadece bir tırnak işaretiyle admin paneli açılıyor!",
			RecommendedStyle: StyleRedVsBlue,
		},
		{
			ID:               "idea-darkweb-leak",
			Category:         "Web, Veri & Kod Güvenliği",
			Title:            "Eski Şifren Şu An Dark Web Listesinde mi?",
			Prompt:           "Veri sızıntısı (Data Breach) ve aynı şifreyi kullanmanın Credential Stuffing riski",
			HookPreview:      "Yıllar önce unuttuğun bir site yüzünden bugünkü bütün hesapların düşebilir!",
			RecommendedStyle: StyleChallenge,
		},
	}
}
