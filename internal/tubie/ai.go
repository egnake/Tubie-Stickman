package tubie

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
	"unicode"
)

const (
	// StyleHybrid balances a cinematic opening hook with clear, actionable cybersecurity takeaways.
	StyleHybrid = "hybrid"
	// StyleCinematic focuses on story-driven tension, realistic dialogue, and narrative atmosphere.
	StyleCinematic = "cinematic"
	// StyleInformative delivers rapid-fire checks, myth-busting facts, and structured defense steps.
	StyleInformative = "informative"
)

const defaultOllamaBaseURL = "http://localhost:11434"

// ScriptResult represents the complete generated script, title, and storyboard.
type ScriptResult struct {
	Title     string            `json:"title"`
	Slug      string            `json:"slug"`
	Style     string            `json:"style,omitempty"`
	Engine    string            `json:"engine,omitempty"`
	Narration string            `json:"narration"`
	Scenes    []SceneDefinition `json:"scenes"`
}

// NormalizeScriptStyle maps arbitrary/UI style values into canonical script styles.
func NormalizeScriptStyle(raw string) string {
	s := strings.ToLower(strings.TrimSpace(raw))
	switch s {
	case StyleCinematic, "sinematik", "hikaye", "story", "gerilim":
		return StyleCinematic
	case StyleInformative, "bilgi", "bilgilendirici", "rehber", "info", "checklist":
		return StyleInformative
	default:
		return StyleHybrid
	}
}

// GenerateScript produces a viral cybersecurity script from a user prompt using the balanced hybrid style.
func GenerateScript(userPrompt string) (*ScriptResult, error) {
	return GenerateScriptWithStyle(userPrompt, StyleHybrid)
}

// GenerateScriptWithStyle produces a viral cybersecurity script tailored to the requested narrative style.
func GenerateScriptWithStyle(userPrompt string, style string) (*ScriptResult, error) {
	trimmed := strings.TrimSpace(userPrompt)
	if trimmed == "" {
		return nil, fmt.Errorf("lütfen bir video konusu veya fikir yazın")
	}

	normalizedStyle := NormalizeScriptStyle(style)
	allowLiveAI := !testing.Testing() && os.Getenv("TUBIE_OFFLINE_AI") != "1"

	// 1. Try Gemini API if user explicitly configured GEMINI_API_KEY
	if allowLiveAI {
		if apiKey := strings.TrimSpace(os.Getenv("GEMINI_API_KEY")); apiKey != "" {
			if res, err := generateViaGemini(apiKey, trimmed, normalizedStyle); err == nil && res != nil && len(res.Scenes) >= 3 {
				return res, nil
			}
		}
	}

	// 2. Try local Ollama if alive and responsive
	if allowLiveAI {
		if res, err := GenerateScriptViaOllamaAt(defaultOllamaBaseURL, trimmed, normalizedStyle); err == nil && res != nil && len(res.Scenes) >= 3 {
			return res, nil
		}
	}

	// 3. Built-in high-craft Turkish cybersecurity scenario engine (instant, offline, zero failure)
	return generateViaTemplatesWithStyle(trimmed, normalizedStyle), nil
}

type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func selectOllamaModel(models []struct {
	Name string `json:"name"`
}) string {
	if override := strings.TrimSpace(os.Getenv("TUBIE_OLLAMA_MODEL")); override != "" {
		return override
	}

	available := make(map[string]string, len(models))
	for _, m := range models {
		available[strings.ToLower(m.Name)] = m.Name
	}

	preferred := []string{
		"qwen2.5:14b",
		"qwen2.5:7b",
		"qwen2.5:latest",
		"gemma2:9b",
		"llama3.1:8b",
		"llama3.2:latest",
		"mistral:latest",
	}
	for _, p := range preferred {
		if orig, ok := available[p]; ok {
			return orig
		}
	}

	// Fallback to any general text-completion model (skipping vision/embed/coder when alternatives exist)
	for _, m := range models {
		low := strings.ToLower(m.Name)
		if strings.Contains(low, "vision") || strings.Contains(low, "llava") || strings.Contains(low, "embed") || strings.Contains(low, "coder") {
			continue
		}
		return m.Name
	}
	if len(models) > 0 {
		return models[0].Name
	}
	return "qwen2.5:7b"
}

func buildScriptSystemPrompt(topic string, style string) string {
	var styleDirective string
	switch style {
	case StyleCinematic:
		styleDirective = `SEÇİLEN TARZ: SİNEMATİK GERİLİM VE HİKÂYE
- İzleyiciyi daha ilk cümlede sahnenin içine sok: telefon titremesi, gelen sahte bir bildirim veya tırnak içinde gerçekçi bir mesaj diyaloğuyla başla.
- Gerilimi adım adım tırmandır: saldırganın arka planda neyi hedeflediğini hissettir, ardından çarpıcı bir kırılma anı ve net bir kurtuluş kuralı ver.
- Kısa, nefesli, sinematik cümleler kullan ("Bir mesaj geliyor… Gönderen tanıdık gibi…").`
	case StyleInformative:
		styleDirective = `SEÇİLEN TARZ: YOĞUN BİLGİ AKIŞI VE REHBER
- İlk 3 saniyede keskin bir test sorusu veya şaşırtıcı bir tespit ile aç ("Sahte bir linki 5 saniyede nasıl anlarsın?").
- Hemen ardından 3-4 somut, uygulanabilir kontrol maddesi ver (küçük harf hileleri, sahte aciliyet, gönderici doğrulaması gibi).
- Mutlaka yaygın bir yanlışı çürüt (ör. "HTTPS kilit simgesi tek başına güven garantisi değildir").
- Net, akılda kalıcı bir aksiyon çağrısıyla bitir.`
	default:
		styleDirective = `SEÇİLEN TARZ: SİNEMATİK + BİLGİ AKIŞI (HİBRİT)
- İLK 3 SANİYE (SİNEMATİK KANCA): İzleyiciyi olayın tam ortasına atan canlı bir sahne ve gerçekçi bir mesaj/olay örneğiyle başla ("Bir mesaj geliyor… Gönderen tanıdık gibi. Logo gerçek… Peki ya bu bir tuzaksa?").
- KÖPRÜ: Tek cümleyle soruna bağla ("İşte saldırılar tam böyle başlıyor! Peki bunu 5 saniyede nasıl fark edersin?").
- BİLGİ AKIŞI: 3-4 somut, teknik olarak doğru ve hemen uygulanabilir kontrol kuralı + bir efsane yıkıcı gerçek ver (ör. kilit simgesi/acil uyarı manipülasyonu gibi).
- FİNAL VURUŞ: Kısa ve vurucu bir uyarı ("DUR. DÜŞÜN. DOĞRULA!") ve akıcı kapanış ("Takipte kal, güvende kal!").`
	}

	return fmt.Sprintf(`Sen Türkiye'nin en yüksek izlenme ve kaydetme oranlarına sahip siber güvenlik YouTube Shorts / Instagram Reels senaristi ve yönetmenisin.

KONU: %q
%s

KALİTE KURALLARI:
1. KESİNLİKLE robotik yapay zeka klişeleri KULLANMA ("konusu sandığından çok daha kritik bir risk taşıyor olabilir", "Hackerlar her gün binlerce kullanıcının dikkatsizliğinden faydalanarak..." gibi kalıplar YASAK).
2. Seslendirme metni 105-155 kelime arasında olsun (35-50 saniyelik hızlı ve doyurucu anlatım).
3. Doğal, konuşma diline uygun, ritmik Türkçe yaz; kısa vurucu cümlelerle ("…", "?", "!") izleyiciyi ekranda tut.
4. Soyut öğüt vermek yerine somut örnek, tırnak içi mesaj örneği veya teknik ipucu ver.
5. Tam olarak 6 sahne planla. Her sahne başlığı kısa (en fazla 4-5 kelime) ve BÜYÜK HARF olmalı.
6. Her sahne için şu varlık yollarından uygun olanı seç:
   "@b1/cyber-laptop.png", "@b1/cyber-skull.png", "@b1/cyber-terminal.png", "@b1/cyber-wifi.png", "@b1/cyber-padlock-unlocked.png", "@b1/cyber-padlock-locked.png", "@b1/cyber-shield.png", "@b1/cyber-bug.png", "@b1/cyber-database.png", "@b1/idea-bulb-crayon.png"

ÖRNEK TON VE KALİTE REFERANSI:
"Bir mesaj geliyor… Gönderen tanıdık gibi. Logo gerçek: 'Güvenliğiniz için hesabınızı hemen doğrulayın!' Her şey normal görünüyor, değil mi? Peki ya o bağlantı bir tuzaksa? İlk olarak internet adresini dikkatlice incele…"

Yanıtını SADECE geçerli JSON formatında ver:
{
  "title": "Dikkat çekici, merak uyandıran Türkçe video başlığı",
  "narration": "105-155 kelimelik yüksek kaliteli Türkçe seslendirme metni",
  "scenes": [
    {"title": "SAHNE 1 BAŞLIĞI", "sfx": "keyboard", "color": "#00F0FF", "prop": {"path": "@b1/cyber-laptop.png"}, "arrow": false},
    {"title": "SAHNE 2 BAŞLIĞI", "sfx": "alert", "color": "#FF3B30", "prop": {"path": "@b1/cyber-skull.png"}, "arrow": true},
    {"title": "SAHNE 3 BAŞLIĞI", "sfx": "glitch", "color": "#FFCC00", "prop": {"path": "@b1/cyber-terminal.png"}, "arrow": false},
    {"title": "SAHNE 4 BAŞLIĞI", "sfx": "paper", "color": "#E056FD", "prop": {"path": "@b1/cyber-padlock-unlocked.png"}, "arrow": false},
    {"title": "SAHNE 5 BAŞLIĞI", "sfx": "lock", "color": "#00FF66", "prop": {"path": "@b1/cyber-shield.png"}, "arrow": false},
    {"title": "SAHNE 6 BAŞLIĞI", "sfx": "chime", "color": "#00F0FF", "prop": {"path": "@b1/cyber-padlock-locked.png"}, "arrow": false}
  ]
}`, topic, styleDirective)
}

// GenerateScriptViaOllamaAt queries an Ollama endpoint (separated for deterministic testing with httptest).
func GenerateScriptViaOllamaAt(baseURL, prompt, style string) (*ScriptResult, error) {
	base := strings.TrimRight(baseURL, "/")
	pingClient := &http.Client{Timeout: 600 * time.Millisecond}
	pingResp, err := pingClient.Get(base + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer pingResp.Body.Close()
	if pingResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama not ready: status %d", pingResp.StatusCode)
	}

	var tags ollamaTagsResponse
	_ = json.NewDecoder(pingResp.Body).Decode(&tags)
	modelName := selectOllamaModel(tags.Models)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	reqBody := map[string]any{
		"model":  modelName,
		"prompt": buildScriptSystemPrompt(prompt, style),
		"stream": false,
		"format": "json",
		"options": map[string]any{
			"temperature": 0.75,
			"top_p":       0.9,
			"num_predict": 1000,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, base+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama status: %d", resp.StatusCode)
	}

	var ollamaResp struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, err
	}

	return parseAndNormalizeScript(ollamaResp.Response, prompt, style, "ollama:"+modelName)
}

func generateViaGemini(apiKey, prompt, style string) (*ScriptResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 18*time.Second)
	defer cancel()

	endpoint := "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent"
	reqBody := map[string]any{
		"contents": []map[string]any{
			{
				"parts": []map[string]string{
					{"text": buildScriptSystemPrompt(prompt, style)},
				},
			},
		},
		"generationConfig": map[string]any{
			"responseMimeType": "application/json",
			"temperature":      0.75,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-goog-api-key", apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gemini status: %d", resp.StatusCode)
	}

	var gemResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&gemResp); err != nil {
		return nil, err
	}
	if len(gemResp.Candidates) == 0 || len(gemResp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("empty gemini response")
	}

	return parseAndNormalizeScript(gemResp.Candidates[0].Content.Parts[0].Text, prompt, style, "gemini")
}

func stripJSONFences(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "```") {
		if idx := strings.Index(trimmed, "\n"); idx != -1 {
			trimmed = trimmed[idx+1:]
		}
		if endIdx := strings.LastIndex(trimmed, "```"); endIdx != -1 {
			trimmed = trimmed[:endIdx]
		}
	}
	start := strings.Index(trimmed, "{")
	end := strings.LastIndex(trimmed, "}")
	if start != -1 && end > start {
		return trimmed[start : end+1]
	}
	return strings.TrimSpace(trimmed)
}

func decodeRawProp(raw json.RawMessage) *SceneProp {
	if len(raw) == 0 || string(raw) == "null" {
		return nil
	}
	var propPath string
	if err := json.Unmarshal(raw, &propPath); err == nil {
		propPath = strings.TrimSpace(propPath)
		if propPath == "" {
			return nil
		}
		return &SceneProp{Path: propPath}
	}
	var obj struct {
		Path  string  `json:"path"`
		Asset string  `json:"asset"`
		X     int     `json:"x"`
		Y     int     `json:"y"`
		W     int     `json:"w"`
		From  float64 `json:"from,omitempty"`
		To    float64 `json:"to,omitempty"`
	}
	if err := json.Unmarshal(raw, &obj); err == nil {
		path := strings.TrimSpace(obj.Path)
		if path == "" {
			path = strings.TrimSpace(obj.Asset)
		}
		if path == "" {
			return nil
		}
		return &SceneProp{
			Path: path,
			X:    obj.X,
			Y:    obj.Y,
			W:    obj.W,
			From: obj.From,
			To:   obj.To,
		}
	}
	return nil
}

func parseAndNormalizeScript(rawJSON, fallbackTopic, style, engine string) (*ScriptResult, error) {
	cleaned := stripJSONFences(rawJSON)
	type rawScene struct {
		Title     string          `json:"title"`
		Color     string          `json:"color"`
		Underline bool            `json:"underline"`
		SFX       string          `json:"sfx"`
		Stickmen  []StickmanActor `json:"stickmen,omitempty"`
		Prop      json.RawMessage `json:"prop,omitempty"`
		Arrow     bool            `json:"arrow"`
	}
	var parsed struct {
		Title     string     `json:"title"`
		Narration string     `json:"narration"`
		Scenes    []rawScene `json:"scenes"`
	}
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		return nil, err
	}

	title := strings.TrimSpace(parsed.Title)
	narration := strings.TrimSpace(parsed.Narration)
	if title == "" {
		title = turkishTitle(fallbackTopic)
	}
	if narration == "" || len(parsed.Scenes) < 3 {
		return nil, fmt.Errorf("incomplete ai script response")
	}

	scenes := make([]SceneDefinition, len(parsed.Scenes))
	for i, rs := range parsed.Scenes {
		scenes[i] = SceneDefinition{
			Title:     rs.Title,
			Color:     rs.Color,
			Underline: rs.Underline,
			SFX:       rs.SFX,
			Stickmen:  rs.Stickmen,
			Prop:      decodeRawProp(rs.Prop),
			Arrow:     rs.Arrow,
		}
	}

	return &ScriptResult{
		Title:     title,
		Slug:      slugify(title),
		Style:     style,
		Engine:    engine,
		Narration: narration,
		Scenes:    normalizeScenes(scenes),
	}, nil
}

var allowedB1Props = map[string]bool{
	"@b1/cyber-laptop.png":           true,
	"@b1/cyber-skull.png":            true,
	"@b1/cyber-terminal.png":         true,
	"@b1/cyber-wifi.png":             true,
	"@b1/cyber-padlock-unlocked.png": true,
	"@b1/cyber-padlock-locked.png":   true,
	"@b1/cyber-shield.png":           true,
	"@b1/cyber-bug.png":              true,
	"@b1/cyber-database.png":         true,
	"@b1/idea-bulb-crayon.png":       true,
	"@b1/idea-note-crayon.png":       true,
}

func normalizeScenes(scenes []SceneDefinition) []SceneDefinition {
	defaultColors := []string{"#00F0FF", "#FF3B30", "#FFCC00", "#E056FD", "#00FF66", "#00F0FF"}
	defaultSFX := []string{"keyboard", "alert", "glitch", "paper", "lock", "chime"}
	defaultProps := []string{
		"@b1/cyber-laptop.png",
		"@b1/cyber-skull.png",
		"@b1/cyber-terminal.png",
		"@b1/idea-bulb-crayon.png",
		"@b1/cyber-shield.png",
		"@b1/cyber-padlock-locked.png",
	}
	validSFX := map[string]bool{
		"keyboard": true, "alert": true, "glitch": true,
		"paper": true, "lock": true, "chime": true, "jingle": true, "tap": true,
	}

	out := make([]SceneDefinition, len(scenes))
	for i, sc := range scenes {
		sc.Title = strings.ToUpperSpecial(unicode.TurkishCase, strings.TrimSpace(sc.Title))
		if sc.Title == "" {
			sc.Title = fmt.Sprintf("SAHNE %d", i+1)
		}
		if !strings.HasPrefix(sc.Color, "#") || len(sc.Color) != 7 {
			sc.Color = defaultColors[i%len(defaultColors)]
		}
		if !validSFX[strings.ToLower(sc.SFX)] {
			sc.SFX = defaultSFX[i%len(defaultSFX)]
		} else {
			sc.SFX = strings.ToLower(sc.SFX)
		}
		if sc.Prop == nil || !allowedB1Props[sc.Prop.Path] {
			sc.Prop = &SceneProp{Path: defaultProps[i%len(defaultProps)], X: 390, Y: 1040, W: 500}
		} else {
			if sc.Prop.X == 0 {
				sc.Prop.X = 390
			}
			if sc.Prop.Y == 0 {
				sc.Prop.Y = 1040
			}
			if sc.Prop.W == 0 {
				sc.Prop.W = 500
			}
		}
		out[i] = sc
	}
	return out
}

type topicBlueprint struct {
	keywords  []string
	titles    map[string]string
	narration map[string]string
	scenes    []SceneDefinition
}

var cyberTopicLibrary = []topicBlueprint{
	{
		keywords: []string{
			"sahte link", "link", "baglanti", "bağlantı", "oltalama", "phishing",
			"tikla", "tıkla", "https", "kilit simge", "e-posta", "mail", "kargo", "dogrula", "doğrula",
		},
		titles: map[string]string{
			StyleHybrid:      "Bu Linke Tıklarsan Ne Olur?",
			StyleCinematic:   "O Tek Tıkla Her Şey Nasıl Gider?",
			StyleInformative: "Sahte Linki 5 Saniyede Yakalama Rehberi",
		},
		narration: map[string]string{
			StyleHybrid:      "Bir mesaj geliyor… Gönderen tanıdık gibi. Logo gerçek. Mesaj oldukça inandırıcı: “Güvenliğiniz için hesabınızı hemen doğrulayın!” Her şey normal görünüyor, değil mi? Peki ya o bağlantı, kişisel bilgilerini ele geçirmek için hazırlanmış bir tuzaksa? İşte oltalama saldırıları tam olarak böyle başlıyor! Peki, sahte bir linki 5 saniyede nasıl anlarsın? İlk olarak internet adresini dikkatlice incele. Küçük harf farklılıkları ve şüpheli alan adları büyük tehlikenin habercisi olabilir. Göndereni kontrol et; tanıdığın bir kurumun adını kullanmaları mesajın gerçekten o kurumdan geldiği anlamına gelmez. Seni acele ettirmeye çalışan “Hemen tıkla”, “Hesabın kapatılacak” gibi ifadeler manipülasyonun bir parçası olabilir. Unutma! HTTPS veya kilit simgesi de tek başına güvenilirlik garantisi değildir. Çünkü bütün dijital hayatını tehlikeye atmak için saldırganların ihtiyacı olan tek şey senin bir anlık dikkatsizliğindir. Tıklamadan önce dur, düşün, doğrula! Güçlü parolalar kullan, iki aşamalı doğrulamayı etkinleştir. Takipte kal, güvende kal!",
			StyleCinematic:   "Telefonun ekranı aydınlanıyor… Tanıdık bir kurum logosu, kusursuz bir tasarım ve seni kalbinden vuran o cümle: “Hesabınızda şüpheli giriş tespit edildi, 15 dakika içinde doğrulayın!” Bir an panikliyorsun. Parmağın o mavi bağlantıya doğru uzanıyor… İşte tam o saniyede dur! Çünkü gördüğün sayfa, gerçeğinin birebir kopyalanmış bir dijital sahnesi. Adres çubuğunda küçücük bir harf değişmiş; üstelik tarayıcıdaki kilit simgesi bile orada duruyor! Sen şifreni yazdığın an, arka planda açılan terminal ekranına bütün oturum anahtarların ve kişisel verilerin akmaya başlıyor. Gerçek kurumlar senden asla panik yaratan bir linkle parola istemez. Bir dahaki sefere acele ettiren bir bildirim gördüğünde bağlantıya değil, uygulamanın kendisine gir. Saldırganların en büyük silahı senin bir anlık dalgınlığın; kalkanın ise durup kontrol etmen. Takipte kal, güvende kal!",
			StyleInformative: "Sahte bir linki sadece beş saniyede nasıl ayıklarsın? İşte oltalama saldırılarını anında ele veren dört net kontrol! Bir: Alan adını harf harf oku. Saldırganlar küçük l yerine büyük ı veya ek tireler kullanır. İki: Görünen ada değil gerçek göndericiye bak; tanıdık logo ve isim taklit edilebilir. Üç: Aciliyet manipülasyonuna düşme. “Hemen tıkla” ya da “Hesabın kapanacak” gibi cümleler mantıklı düşünmeni engellemek için yazılır. Dört: Adres çubuğundaki HTTPS ve kilit simgesi sitenin dürüst olduğunu değil, yalnızca bağlantının şifreli olduğunu gösterir; sahte sitelerde de kilit simgesi vardır! Şüpheli bir mesaj aldığında linke asla dokunma, doğrudan resmi uygulama veya tarayıcıdan kendin giriş yap. İki aşamalı doğrulamayı aktif tut. Bu kuralları kaydet, güvende kal!",
		},
		scenes: []SceneDefinition{
			{Title: "SAHTE MESAJ EKRANDA!", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 390, Y: 1040, W: 500}},
			{Title: "ACELE DOĞRULAMA TUZAĞI!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}, Arrow: true},
			{Title: "KÜÇÜK HARF VE ALAN ADI HİLESİ", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}},
			{Title: "KİLİT SİMGESİ GARANTİ DEĞİL!", Color: "#E056FD", SFX: "paper", Prop: &SceneProp{Path: "@b1/cyber-padlock-unlocked.png", X: 390, Y: 1040, W: 500}},
			{Title: "DUR. DÜŞÜN. DOĞRULA!", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}},
			{Title: "2FA AÇ, GÜVENDE KAL!", Color: "#00F0FF", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 390, Y: 1040, W: 500}},
		},
	},
	{
		keywords: []string{"instagram", "insta", "sosyal medya", "hesap cal", "hesap çal", "mavi tik", "telif", "dm", "tiktok"},
		titles: map[string]string{
			StyleHybrid:      "İnstagram Hesabın Tek DM İle Nasıl Gider?",
			StyleCinematic:   "O DM'ye Tıkladığın Gece Hesabın Nasıl Gider?",
			StyleInformative: "İnstagram Hesap Çalma Tuzaklarından Korunma",
		},
		narration: map[string]string{
			StyleHybrid:      "Gece telefonuna bir DM düşüyor… Mesaj resmi görünüyor: “Telif hakkı ihlali nedeniyle hesabınız 24 saat içinde kapatılacak, itiraz için tıklayın!” Kalbin hızlanıyor, değil mi? İşte yıllarca büyüttüğün Instagram hesabını ele geçirmek için kurulan tuzak tam olarak böyle başlıyor! Hackerlar artık şifre kırmakla uğraşmıyor; sahte giriş ekranıyla şifreni senin yazmanı bekliyor. Peki kendini saniyeler içinde nasıl korursun? Bir: Instagram yetkilileri asla DM üzerinden form veya şifre bağlantısı göndermez. İki: Gelen mesajların gerçekliğini sadece Ayarlar içindeki “Instagram E-postaları” bölümünden kontrol et. Üç: SMS yerine mutlaka Google Authenticator gibi bir doğrulama uygulaması kullan; çünkü SIM kopyalama veya sahte formlara karşı en sağlam kalkan budur! Unutma, bir anlık panik bütün emeğini sıfırlayabilir. Şüpheli DM’leri sil, iki faktörlü korumayı hemen aç. Takipte kal, güvende kal!",
			StyleCinematic:   "Saat gece yarısı. Yıllardır emek verdiğin hesabına tek bir bildirim geliyor: “Profiliniz mavi tik doğrulamasına seçildi” ya da “Telif ihlali nedeniyle hesabınız silinecek.” Linke bastığında karşına Instagram'ın birebir aynısı bir giriş ekranı çıkıyor. Kullanıcı adını ve şifreni yazdığın saniye ekran donuyor… Arka planda ise saldırgan şifreni, e-postanı ve kurtarma kodlarını çoktan değiştirmiş oluyor! Sabah uyandığında profil fotoğrafın değişmiş, takipçilerine kripto dolandırıcılığı mesajları gidiyor. Bu kâbusu yaşamamanın tek yolu o ilk panik anını yenmek. Resmi bildirimlerin tek adresi ayarlar menüsündeki güvenlik e-postalarıdır. DM’den gelen hiçbir doğrulama linkine dokunma, Authenticator tabanlı iki faktörlü korumayı mutlaka aç. Emeğini bir tıkla hediye etme!",
			StyleInformative: "Instagram hesabını çalınmaktan koruyacak dört kesin kural! Birinci kural: “Telif ihlali” veya “Mavi tik kazandınız” diyen DM’lere asla tıklama; Instagram destek ekibi DM’den link atmaz. İkinci kural: Şüpheli bir mail veya mesaj geldiğinde doğrulamayı linkten değil, Ayarlar > Hesaplar Merkezi > Parola ve Güvenlik > Son E-postalar sekmesinden yap. Üçüncü kural: İki faktörlü doğrulamayı SMS yerine mutlaka Authenticator uygulamasıyla aç ve yedek kodları çevrimdışı sakla. Dördüncü kural: Hesabına bağlı bilmediğin üçüncü parti uygulamaları ve eski giriş oturumlarını hemen temizle. Linke yanlışlıkla şifre girdiysen saniyeler içinde parolanı değiştir ve tüm cihazlardan çıkış yap. Bu rehberi kaydet, arkadaşlarını da uyar!",
		},
		scenes: []SceneDefinition{
			{Title: "SAHTE DM BİLDİRİMİ!", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 390, Y: 1040, W: 500}},
			{Title: "TELİF VE MAVİ TİK TUZAĞI!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}, Arrow: true},
			{Title: "OTURUM VE ŞİFRE ELE GEÇTİ", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}},
			{Title: "INSTAGRAM DM'DEN ŞİFRE İSTEMEZ", Color: "#E056FD", SFX: "paper", Prop: &SceneProp{Path: "@b1/idea-bulb-crayon.png", X: 390, Y: 1040, W: 500}},
			{Title: "AUTHENTICATOR İLE 2FA AÇ!", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}},
			{Title: "EMEĞİNİ VE HESABINI KORU!", Color: "#00F0FF", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 390, Y: 1040, W: 500}},
		},
	},
	{
		keywords: []string{"kripto", "bitcoin", "cuzdan", "cüzdan", "airdrop", "token", "coin", "nft", "seed", "metamask"},
		titles: map[string]string{
			StyleHybrid:      "Kripto Cüzdanın Tek Onayla Nasıl Boşaltılır?",
			StyleCinematic:   "Bir Airdrop İmzasında Tüm Cüzdan Nasıl Gitti?",
			StyleInformative: "Kripto Cüzdanını Boşaltan 3 Akıllı Tuzak",
		},
		narration: map[string]string{
			StyleHybrid:      "Cüzdanına durduk yere değerli bir token veya özel airdrop daveti düşüyor… “Ödülünü almak için cüzdanını bağla!” diyor. Heyecanla cüzdanı bağlayıp o küçük imza onayını veriyorsun. İşte o saniye tüm birikimin tek bir blokta buharlaşıyor! Hackerlar bunun için gizli kelimelerini istemek zorunda bile değil; verdiğin sınırsız harcama onayı yani Approve izniyle cüzdanındaki varlıkları kendi adreslerine çekiyorlar. Peki cüzdanını nasıl korursun? Bir: Kimse durduk yere bedava para dağıtmaz; tanımadığın airdrop sitelerine asla ana cüzdanını bağlama. İki: İmza ekranında “SetApprovalForAll” veya sınırsız harcama izni görürsen anında reddet. Üç: Büyük bakiyeni günlük kullandığın sıcak cüzdandan ayırıp soğuk cüzdanda sakla. Onaylarını düzenli olarak kontrol et ve iptal et. Parana sahip çık, güvende kal!",
			StyleCinematic:   "Ekranda parlak bir ödül sayfası açılıyor: “Erken kullanıcı olduğunuz için bin dolarlık airdrop kazandınız!” Tek yapman gereken cüzdanını bağlayıp işlemi onaylamak. Onayla butonuna bastığın an her şey normal görünüyor… Ta ki beş saniye sonra bakiyenin sıfırlandığını görene kadar! Saldırgan senin gizli on iki kelimeni çalmadı; ona bizzat akıllı sözleşme üzerinden sınırsız çekim yetkisi verdin! Kripto dünyasında en tehlikeli tuzaklar artık şifre istemez, sadece aceleyle verilen tek bir imzaya bakar. Bu yüzden büyük varlıklarını günlük test ettiğin sıcak cüzdanlardan kesinlikle ayır. Bilinmeyen her airdrop bağlantısına şüpheyle yaklaş, verdiğin tüm sözleşme izinlerini düzenli temizle!",
			StyleInformative: "Kripto cüzdanını sıfırlayan üç gizli tuzağı bilmeden işlem onaylama! Bir: Cüzdana gelen sahte NFT ve tokenlar. Üzerindeki web adresine girip cüzdan bağlarsan kötü amaçlı sözleşmeye izin vermiş olursun. İki: Sınırsız harcama yetkisi. İşlem onay penceresinde “Approve” veya “SetApprovalForAll” yetkisi isteyen her site varlıklarını çekebilir. Üç: Sahte destek ekipleri. Telegram veya X üzerinden sana önce yazıp “Cüzdanını senkronize et” diyen herkes dolandırıcıdır. Korunmak için ana bakiyeni soğuk cüzdanda tut, deneme siteleri için ayrı boş bir cüzdan kullan ve gizli kurtarma kelimelerini asla dijital ortama yazma!",
		},
		scenes: []SceneDefinition{
			{Title: "BEDAVA AİRDROP TUZAĞI!", Color: "#FFCC00", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 390, Y: 1040, W: 500}},
			{Title: "SINIRSIZ HARCAMA İZNİ!", Color: "#FF3B30", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}, Arrow: true},
			{Title: "CÜZDAN BAKİYESİ SIFIRLANDI!", Color: "#FF3B30", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}},
			{Title: "GİZLİ KELİME BİLE GEREKMİYOR!", Color: "#00F0FF", SFX: "paper", Prop: &SceneProp{Path: "@b1/idea-bulb-crayon.png", X: 390, Y: 1040, W: 500}},
			{Title: "SOĞUK CÜZDAN VE İZİN İPTALİ", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 390, Y: 1040, W: 500}},
			{Title: "VARLIKLARINI KORU, UYANIK OL!", Color: "#00FF66", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}},
		},
	},
	{
		keywords: []string{"whatsapp", "sms", "onay kodu", "dogrulama kodu", "doğrulama kodu", "arkadas", "arkadaş"},
		titles: map[string]string{
			StyleHybrid:      "WhatsApp Hesabın Tek SMS Koduyla Nasıl Gider?",
			StyleCinematic:   "En Yakın Arkadaşından Gelen Mesaj Tuzağı!",
			StyleInformative: "WhatsApp SMS Kodu Hırsızlığını Önleme Rehberi",
		},
		narration: map[string]string{
			StyleHybrid:      "Telefonuna resmi WhatsApp numarasından altı haneli bir SMS onay kodu geliyor. Tam “Bu ne?” derken en yakın arkadaşından mesaj düşüyor: “Kanka kodu yanlışlıkla senin numarana yollamışım, acil atar mısın?” Tanıdık biri olduğu için tereddüt etmeden yazıyorsun… Ve üç saniye sonra WhatsApp seni oturumdan atıyor! Çünkü arkadaşının hesabı zaten ele geçirilmişti; saldırgan senin numaranla yeni bir cihazdan giriş yapıp senden o tek kullanımlık anahtarı aldı! Şimdi tüm rehberine senin adınla acil para isteme mesajları gidiyor. Peki bu tuzağı nasıl bozarsın? Bir: SMS doğrulama kodunu annene bile mesajla paylaşma. İki: Arkadaşından garip bir talep gelirse mutlaka sesli arayarak doğrula. Üç: WhatsApp Ayarlar > Hesap > İki Adımlı Doğrulama bölümünden altı haneli kişisel PIN kodunu hemen aç! Kodunu koru, çevreni koru!",
			StyleCinematic:   "Öğle arası ekranına sıradan bir SMS düşüyor: “WhatsApp kodunuz: 482 910. Kimseyle paylaşmayın.” Tam umursamayacakken en samimi arkadaşından aceleci bir mesaj geliyor: “Telefonum sıfırlandı, yanlışlıkla senin numaranı yazmışım, kodu yollar mısın?” İyi niyetle o altı rakamı gönderdiğin saniye ekran kararıyor… “Telefon numaranız artık bu cihazda kayıtlı değil.” Artık senin profil fotoğrafınla arkadaşlarına, ailene, iş çevrene borç para isteyen mesajlar akmaya başlıyor! Gerçek şu: O kodu hiçbir zaman arkadaşın istemedi, ele geçirilmiş bir hesabın arkasındaki dolandırıcı istedi. Şüpheli bir istekte mutlaka sesli arama yap ve iki adımlı doğrulama PIN’ini bugünden aktif et!",
			StyleInformative: "WhatsApp hesabının tek bir mesajla çalınmasını engellemek için üç net önlem! Bir: Altı haneli SMS doğrulama kodu hesabının anahtarıdır; kim yanlışlıkla istediğini söylerse söylesin asla paylaşma. İki: Tanıdığın birinden kod ya da acil para talebi gelirse yazışmaya devam etmeden doğrudan normal telefon aramasıyla teyit et. Üç: WhatsApp Ayarlar > Hesap > İki Adımlı Doğrulama kısmından kendi altı haneli PIN kodunu oluştur ve kurtarma e-postanı ekle. Böylece saldırgan SMS kodunu ele geçirse bile senin belirlediğin PIN olmadan hesabına asla giremez. Bu bilgiyi ailene de gönder!",
		},
		scenes: []SceneDefinition{
			{Title: "BEKLENMEDİK 6 HANELİ SMS KODU", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-wifi.png", X: 390, Y: 1040, W: 500}},
			{Title: "ARKADAŞINDAN SAHTE MESAJ!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 390, Y: 1040, W: 500}, Arrow: true},
			{Title: "KODU VERDİĞİN AN HESAP GİDER!", Color: "#FF3B30", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}},
			{Title: "REHBERİNE DOLANDIRICILIK MESAJI", Color: "#FFCC00", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}},
			{Title: "İKİ ADIMLI DOĞRULAMA PİN'İ AÇ", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}},
			{Title: "KODU ASLA KİMSEYE VERME!", Color: "#00FF66", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 390, Y: 1040, W: 500}},
		},
	},
	{
		keywords: []string{"wifi", "wi-fi", "kafe", "ortak ag", "ortak ağ", "hotspot", "havalimani", "havalimanı", "vpn", "deauth", "evil twin"},
		titles: map[string]string{
			StyleHybrid:      "Kafe Wi-Fi Ağına Bağlanınca Arkada Ne Olur?",
			StyleCinematic:   "Ücretsiz Kafe Wi-Fi Tuzağı: Görünmeyen Köprü",
			StyleInformative: "Ortak Wi-Fi Ağlarında Güvende Kalmanın 4 Yolu",
		},
		narration: map[string]string{
			StyleHybrid:      "Bir kafeye oturuyorsun, şifreyi sorup hemen ücretsiz Wi-Fi ağına bağlanıyorsun. Rahatça e-postalarına ve bankana bakıyorsun, değil mi? Peki ya bağlandığın ağ gerçekten kafenin değilse? Saldırganlar hemen yan masadan birebir aynı isimle “Evil Twin” yani sahte ikiz bir ağ açabilir! Telefonun otomatik olarak en güçlü sinyale atladığında bütün internet trafiğin artık onların bilgisayarından akmaya başlar. Hatta Deauth saldırısıyla seni gerçek ağdan düşürüp şifreli paketleri havada yakalayabilirler! Peki dışarıda nasıl güvende kalırsın? Bir: Ortak Wi-Fi üzerinde asla bankacılık ve kritik giriş işlemi yapma; bunun için mobil verini kullan. İki: Telefonunun “Bilinen ağlara otomatik bağlan” ayarını mutlaka kapat. Üç: Bağlanmak zorundaysan trafiğini şifreleyen güvenilir bir VPN aç! Ücretsiz internet pahalıya patlamasın. Takipte kal, güvende kal!",
			StyleCinematic:   "Kahveni alıp masaya oturdun ve kafenin adıyla birebir aynı olan o ücretsiz Wi-Fi ağına bağlandın. Her şey sorunsuz çalışıyor… Ama iki masa ötedeki dizüstü bilgisayarda senin cihazından çıkan her istek anlık olarak akıyor! Çünkü bağlandığın şey kafenin yönlendiricisi değil; saldırganın kendi bilgisayarından yayınladığı sahte bir ikiz ağ. Gerçek ağı kısa bir kesinti saldırısıyla düşürdü ve senin telefonunu kendi tuzağına çekti. Şifresiz trafikler, oturum çerezleri, girmeye çalıştığın paneller tek bir ekranda toplanıyor. Ücretsiz bir Wi-Fi için dijital güvenliğini riske atmaya değer mi? Dışarıda hassas işlemler için kendi hücresel verini kullan, otomatik Wi-Fi bağlanmayı kapat ve VPN kalkanını devrede tut!",
			StyleInformative: "Halka açık Wi-Fi ağlarında verilerinin izlenmesini önleyecek dört pratik adım! Bir: Bankacılık, e-devlet ve ana hesap girişlerini ortak Wi-Fi yerine daima kendi mobil internetinden yap. İki: Telefon ayarlarından “Ağlara otomatik bağlan” seçeneğini kapat; böylece cihazın aynı isimdeki sahte ikiz ağlara habersizce atlamaz. Üç: Herkese açık ağları kullanmak zorundaysan uçtan uca şifreli güvenilir bir VPN aç. Dört: Tarayıcında yalnızca HTTPS bağlantıları kabul eden güvenlik ayarını aktif tut ve işin bitince ortak ağı unut. Basit alışkanlıklar büyük saldırıları durdurur!",
		},
		scenes: []SceneDefinition{
			{Title: "ÜCRETSİZ KAFE WI-FI AĞI!", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-wifi.png", X: 390, Y: 1040, W: 500}},
			{Title: "SAHTE İKİZ AĞ (EVİL TWİN)!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}, Arrow: true},
			{Title: "TRAFİĞİN YAN MASADAN İZLENİYOR", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}},
			{Title: "OTOMATİK BAĞLANMAYI KAPAT!", Color: "#E056FD", SFX: "paper", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 390, Y: 1040, W: 500}},
			{Title: "VPN VEYA MOBİL VERİ KULLAN", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}},
			{Title: "DİJİTAL İZİNİ ŞİFRELE!", Color: "#00F0FF", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 390, Y: 1040, W: 500}},
		},
	},
	{
		keywords: []string{"şifre", "sifre", "parola", "brute", "kaba kuvvet", "sizinti", "sızıntı", "123456", "passkey"},
		titles: map[string]string{
			StyleHybrid:      "Şifren Saniyeler İçinde Nasıl Kırılıyor?",
			StyleCinematic:   "Aynı Şifre Yüzünden Bütün Hesaplar Nasıl Gitti?",
			StyleInformative: "Kırılması Yüzyıllar Sürecek Parola Formülü",
		},
		narration: map[string]string{
			StyleHybrid:      "“Benim şifrem zaten uzun, kimse tahmin edemez” diyorsun, değil mi? Peki yıllar önce kayıt olduğun sıradan bir forum sitesi hacklendiğinde ne oluyor? Saldırganlar o siteden sızan e-posta ve şifre listesini alıp Instagram’dan bankana kadar yüzlerce platformda otomatik olarak deniyor! Buna kimlik bilgisi doldurma saldırısı denir. Üstelik sekiz karakterli basit şifreler modern ekran kartlarıyla saniyeler içinde kırılıyor. Peki kırılmaz bir savunma nasıl kurulur? Bir: Her hesap için mutlaka birbirinden tamamen farklı bir parola kullan. İki: Tahmin edilmesi kolay kelimeler yerine en az on dört karakterli rastgele bir parola cümlesi oluştur. Üç: Tüm bunları aklında tutmaya çalışma, güvenilir bir parola yöneticisi ve mümkün olan yerlerde Passkey kullan! Bir sızıntı bütün hayatını açmasın. Takipte kal, güvende kal!",
			StyleCinematic:   "Yıllar önce tek bir alışveriş sitesine girdiğin o eski şifre, bugün karanlık ağdaki bir sızıntı dosyasında açık metin olarak duruyor. Saldırganın senin şifreni kırmasına bile gerek yok; otomatik bir bot o şifreyi saniyeler içinde mailine, sosyal medyana ve finans uygulamalarına yazıp deniyor. Aynı şifreyi üç yerde kullandıysan ilk kırılan kapı bütün evin anahtarını veriyor! Çözüm ise karmaşık formüller ezberlemek değil; her kapıya bambaşka bir kilit takmak. Parola yöneticisiyle her siteye eşsiz ve uzun şifreler üret, kritik hesaplarını iki aşamalı doğrulama ve Passkey ile mühürle!",
			StyleInformative: "Hesaplarının saniyeler içinde kırılmasını önleyen üç temel parola gerçeği! Bir: Sekiz haneli ve sadece küçük harfli bir şifre modern donanımlarla dakikalar değil saniyeler içinde çözülür; en az on dört karakter hedeflenmelidir. İki: En büyük hata dünyanın en güçlü şifresini bile beş farklı sitede kullanmaktır; tek bir sitenin veri sızıntısı bütün hesaplarını düşürür. Üç: Büyük harf, rakam ve sembol içeren eşsiz parolaları güvenli bir parola yöneticisinde sakla ve destekleyen platformlarda şifre yerine Passkey aç. Parola hijyenini bugün yenile!",
		},
		scenes: []SceneDefinition{
			{Title: "AYNI ŞİFRE = TEK ANAHTAR RİSKİ", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-database.png", X: 390, Y: 1040, W: 500}},
			{Title: "SANİYEDE MİLYARLARCA DENEME!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}, Arrow: true},
			{Title: "TEK SIZINTI TÜM HESAPLARI AÇAR", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}},
			{Title: "EN AZ 14 KARAKTER KURALI", Color: "#E056FD", SFX: "paper", Prop: &SceneProp{Path: "@b1/idea-bulb-crayon.png", X: 390, Y: 1040, W: 500}},
			{Title: "PAROLA YÖNETİCİSİ VE PASSKEY", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}},
			{Title: "HER HESABA EŞSİZ KİLİT!", Color: "#00F0FF", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 390, Y: 1040, W: 500}},
		},
	},
	{
		keywords: []string{"deepfake", "yapay zeka", "ses klon", "sahte ses", "sahte video", "sahte arama"},
		titles: map[string]string{
			StyleHybrid:      "3 Saniyelik Sesinle Nasıl Dolandırılırsın?",
			StyleCinematic:   "Arayan Gerçekten Yakının mı, Yapay Zeka mı?",
			StyleInformative: "Yapay Zeka Ses ve Deepfake Tuzaklarını Tanı!",
		},
		narration: map[string]string{
			StyleHybrid:      "Gece yarısı telefonun çalıyor… Arayan kişinin sesi birebir kardeşine ya da yakınına benziyor: “Kaza yaptım, acil şu hesaba para at!” Ses tonu, nefesi, paniği tamamen gerçek gibi, değil mi? Ama karşındaki kişi gerçek yakının olmayabilir! Yapay zeka araçları sosyal medyada paylaştığın yalnızca üç beş saniyelik bir videodan sesini birebir klonlayabiliyor. Üstelik sahte aciliyet yaratarak soru sormanı engelliyorlar. Peki bu yeni nesil tuzağı nasıl bozarsın? Bir: Telefonda acele para veya kod isteyen biri ses tanıdık gelse bile hemen telefonu kapat ve kişinin kayıtlı numarasını kendin geri ara. İki: Ailenle önceden yalnızca sizin bildiğiniz gizli bir doğrulama kelimesi belirle! Üç: Sosyal medyada ses ve yüz netliği yüksek videoları herkese açık paylaşırken iki kez düşün. Duyduğun her sese inanmadan önce doğrula. Takipte kal, güvende kal!",
			StyleCinematic:   "Saat gece iki. Ekranda yabancı bir numara, ama hattın ucunda kalbini durduran tanıdık bir ses: ağlayarak acil yardım ve para istiyor. Tek bir saniye bile şüphe etmeden bankacılık uygulamasını açmaya hazırlanıyorsun… Oysa o ses, iki hafta önce paylaştığın kısa bir videodan yapay zekayla beş saniyede kopyalanmış bir ses modeli! Dolandırıcı sadece o panik anında soru sormamanı ve bildiğin numaradan geri arama yapmamanı istiyor. İşte tam o an dur, telefonu kapat ve yakınının gerçek numarasını kendin çevir. Ailenle gizli bir parola belirle ve ses klonlama tuzağına düşme!",
			StyleInformative: "Yapay zeka ses klonlama ve deepfake dolandırıcılığına karşı üç altın kural! Bir: Arayan numara gizliyse veya yabancıysa ve karşıdaki ses yakınınmış gibi acil para istiyorsa görüşmeyi derhal kapatıp gerçek numarayı kendin ara. İki: Ailen ve yakın çevrenle acil durumlar için telefonda sorulacak gizli bir aile parolası belirle; yapay zeka bu kelimeyi tahmin edemez. Üç: Görüntülü aramalarda elini yüzünün önünden geçirmesini veya başını yana çevirmesini isteyerek sahte filtreleri test et. Panik yok, teyit var!",
		},
		scenes: []SceneDefinition{
			{Title: "TANIDIK SESLE PANİK ARAMASI!", Color: "#00F0FF", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-wifi.png", X: 390, Y: 1040, W: 500}},
			{Title: "3 SANİYELİK VİDEODAN SES KLONU!", Color: "#FF3B30", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}, Arrow: true},
			{Title: "ACİL PARA MANİPÜLASYONU", Color: "#FFCC00", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}},
			{Title: "GERÇEK NUMARAYI KENDİN ARA!", Color: "#E056FD", SFX: "paper", Prop: &SceneProp{Path: "@b1/idea-bulb-crayon.png", X: 390, Y: 1040, W: 500}},
			{Title: "AİLE İÇİ GİZLİ PAROLA BELİRLE", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}},
			{Title: "DUYMADAN ÖNCE DOĞRULA!", Color: "#00F0FF", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 390, Y: 1040, W: 500}},
		},
	},
	{
		keywords: []string{"crack", "apk", "dosya", "virus", "virüs", "trojan", "fidye", "ransomware", "exe", "indir"},
		titles: map[string]string{
			StyleHybrid:      "Ücretsiz Crack İndirirken Bilgisayarında Ne Olur?",
			StyleCinematic:   "O Tek Dosyaya Çift Tıkladığın Saniye!",
			StyleInformative: "Zararlı Dosya ve Infostealer Tehlikesinden Korunma",
		},
		narration: map[string]string{
			StyleHybrid:      "Ücretli bir programı veya popüler bir oyunu bedava indirmek için o siteye giriyorsun. “Antivirüsü kapat, kurulumu çalıştır” yazısını görüp tamam diyorsun… İşte o iki saniyede bilgisayarın artık sadece sana ait değil! Günümüzde virüsler ekranını kilitleyip kendini belli etmek zorunda değil. Sessiz bilgi hırsızları yani Infostealer yazılımlar, sen fark etmeden tarayıcında kayıtlı tüm şifrelerini, açık oturum çerezlerini ve cüzdan dosyalarını üç saniyede paketleyip dışarı sızdırır! Peki kendini nasıl korursun? Bir: Antivirüsü kapatmanı isteyen hiçbir kurulum dosyasını asla çalıştırma. İki: Şüpheli dosyaları açmadan önce mutlaka VirusTotal üzerinden tarat veya resmi mağaza dışı APK indirme. Üç: Tarayıcında kritik banka ve hesap şifrelerini açıkta bırakma, iki aşamalı doğrulamayı aktif tut! Bedava görünen tuzaklara tüm verini ödeme. Takipte kal, güvende kal!",
			StyleCinematic:   "İndirme tamamlandı. Klasörü açıp kurulum dosyasına çift tıklıyorsun. Siyah bir komut penceresi yarım saniye görünüp kayboluyor ve hiçbir şey olmamış gibi kurulum başlıyor… Ama o yarım saniyede arka planda sessiz bir bilgi hırsızı çalıştı! Masaüstün silinmedi, bilgisayarın yavaşlamadı; sadece Chrome ve Edge içinde kayıtlı bütün parolaların, Instagram ve Gmail oturum çerezlerin uzaktaki bir sunucuya kopyalandı! Saldırgan artık şifreni bilmese bile aktif oturumunu devralabiliyor. Antivirüsü kapatmanı isteyen her dosya kırmızı alarmdır. Resmi kaynaklardan şaşma, şüpheli tek bir dosyayı bile doğrudan ana sisteminde açma!",
			StyleInformative: "Zararlı yazılım ve bilgi hırsızı tuzaklarını durduran dört net savunma! Bir: “Kurulum için Defender’ı veya antivirüsü kapatın” uyarısı yüzde yüz zararlı yazılım işaretidir; dosyayı anında sil. İki: Üçüncü parti sitelerden crackli program ve modlu APK indirme; çoğunun içinde sessiz infostealer gizlidir. Üç: Kaynağından emin olmadığın PDF, ZIP veya EXE dosyalarını açmadan önce VirusTotal’de tara. Dört: Tarayıcı şifre yöneticisi yerine ana parolayla korunan ayrı bir parola yöneticisi kullan ve önemli dosyalarını düzenli çevrimdışı yedekle!",
		},
		scenes: []SceneDefinition{
			{Title: "BEDAVA CRACK VE APK TUZAĞI!", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 390, Y: 1040, W: 500}},
			{Title: "ANTİVİRÜSÜ KAPAT YALANI!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-bug.png", X: 390, Y: 1040, W: 500}, Arrow: true},
			{Title: "ÇEREZ VE ŞİFRELER 3 SANİYEDE ÇALINDI", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}},
			{Title: "GÖRÜNMEZ İNFOSTEALER TEHLİKESİ", Color: "#E056FD", SFX: "paper", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}},
			{Title: "VİRUSTOTAL VE RESMİ KAYNAK KULLAN", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}},
			{Title: "SİSTEMİNİ TEMİZ TUT, GÜVENDE KAL!", Color: "#00F0FF", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 390, Y: 1040, W: 500}},
		},
	},
	{
		keywords: []string{"sql", "injection", "veritabani", "veritabanı", "xss", "arka kapi", "arka kapı", "api"},
		titles: map[string]string{
			StyleHybrid:      "Tek Bir Tırnak İşaretiyle Veritabanı Nasıl Düşer?",
			StyleCinematic:   "Giriş Formundaki Tek Karakter Tüm Sistemi Açtı!",
			StyleInformative: "SQL Injection Nasıl Çalışır ve Nasıl Önlenir?",
		},
		narration: map[string]string{
			StyleHybrid:      "Sıradan bir web sitesi giriş ekranı düşün… Kullanıcı adı kutusuna normal bir isim yerine tek bir tırnak işareti ve “veya bir eşittir bir” komutu yazılıyor. Ve sistem şifre sormadan doğrudan yönetici panelini açıyor! Nasıl mı? Buna SQL Injection denir. Arka plandaki kod, kullanıcıdan gelen yazıyı veri olarak değil doğrudan veritabanı komutu olarak çalıştırdığında kapılar ardına kadar açılır! Saldırgan tek sorguda tüm kullanıcı tablolarını ve kayıtları çekebilir. Peki geliştiriciler bunu nasıl durdurur? Bir: Asla kullanıcı girdisini string birleştirme ile SQL sorgusuna yapıştırma; her zaman parametrik sorgular yani Prepared Statements kullan. İki: Veritabanı kullanıcısına sadece gereken en düşük yetkiyi ver. Üç: Hata mesajlarını kullanıcı ekranında açık gösterme! Güvenli kod sadece çalışmakla kalmaz, veriyi korur. Takipte kal, güvende kal!",
			StyleCinematic:   "Karanlık bir terminal penceresinde test edilen sıradan bir giriş kutusu… Saldırgan parola bilmiyor, kaba kuvvet bile denemiyor. Sadece kullanıcı adı alanına bir tek tırnak ve kısa bir mantık ifadesi bırakıp enter’a basıyor. Sunucu bu metni sıradan bir isim sanmak yerine kendi sorgusunun parçası olarak çalıştırıyor! Sonuç? Kontrol tamamen atlanıyor, admin oturumu açılıyor ve yıllarca biriken veritabanı tabloları saniyeler içinde listeleniyor. Tüm bunlar tek bir eksik doğrulama satırı yüzünden yaşandı. Kod yazarken kural basit: Kullanıcıdan gelen hiçbir girdiye güvenme, daima parametrik sorgu kullan!",
			StyleInformative: "SQL Injection açığının mantığı ve üç kesin çözümü! Sorun şu: Uygulama kullanıcı girdisini temizlemeden doğrudan SQL metnine eklediğinde, saldırgan sorgunun yapısını değiştirip yetkisiz veri çekebilir. Çözüm bir: Bütün veritabanı işlemlerinde mutlaka parametrik sorgular veya güvenli ORM binding kullan; metin birleştirme yapma. Çözüm iki: En az yetki ilkesini uygula; web uygulamasının veritabanı hesabı tablo silme veya tam okuma yetkisine sahip olmasın. Çözüm üç: Ham SQL hata çıktılarını gizle ve WAF ile anormal sorgu denemelerini izle. Sağlam mimari ilk satırda başlar!",
		},
		scenes: []SceneDefinition{
			{Title: "MASUM GÖRÜNEN GİRİŞ KUTUSU", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 390, Y: 1040, W: 500}},
			{Title: "TEK TIRNAK VE 'OR 1=1' HİLESİ!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}, Arrow: true},
			{Title: "VERİTABANI KONTROLÜ ATLANDI!", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-database.png", X: 390, Y: 1040, W: 500}},
			{Title: "ADMİN YETKİSİ ELE GEÇİRİLDİ!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}},
			{Title: "PARAMETRİK SORGU KULLAN!", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}},
			{Title: "KODUNU KORU, VERİNİ KORU!", Color: "#00F0FF", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 390, Y: 1040, W: 500}},
		},
	},
	{
		keywords: []string{"qr", "quishing", "usb", "sarj", "şarj", "casus", "kamera", "mikrofon", "izin"},
		titles: map[string]string{
			StyleHybrid:      "Sokaktaki QR Kodu Okutmadan Önce Dur!",
			StyleCinematic:   "Otopark QR Koduna Gizlenen Sahte Ödeme Tuzağı!",
			StyleInformative: "QR Kod (Quishing) ve Cihaz İzin Güvenliği",
		},
		narration: map[string]string{
			StyleHybrid:      "Bir restoranda ya da otoparkta ödeme yapmak için hemen kameranı çıkarıp masadaki QR koda tutuyorsun, değil mi? Peki ya o karekodun üstüne geceden başka bir etiket yapıştırıldıysa? Buna Quishing yani QR oltalama saldırısı denir! Gözünle baktığında iki karekodu birbirinden ayıramazsın; ama okuttuğun an açılan sayfa, seni birebir kopyalanmış sahte bir ödeme ya da giriş ekranına yönlendirir. Aynı risk halka açık yerlerdeki bilinmeyen USB şarj kabloları için de geçerlidir! Peki nasıl korunursun? Bir: QR kodu okuttuğunda bağlantıyı doğrudan açma; önce ekranda çıkan alan adını mutlaka oku. İki: Kart ya da şifre isteyen ödemeleri tarayıcı linkinden değil, resmi uygulamadan yap. Üç: Kamusal USB portlarına doğrudan kablo takmak yerine kendi adaptörünü veya powerbankini kullan! Gözünle değil, kontrolle güven. Takipte kal, güvende kal!",
			StyleCinematic:   "Aceleyle otoparktan çıkarken tabeladaki QR kodu okutuyorsun. Ekranda tanıdık bir ödeme sayfası açılıyor, plakanı ve kart bilgini girmek üzeresin… Ama fark etmediğin küçük bir detay var: gerçek QR kodun tam üzerine milimetrik sahte bir çıkartma yapıştırılmış! Kart bilgilerini yazdığın saniye ödeme onay kodu dolandırıcının ekranına düşüyor. Gözümüz karekodun içinde hangi adresin gizlendiğini okuyamaz; ama telefonumuzun link önizlemesi okuyabilir! Okutmadan önce üstündeki etiketi kontrol et, açılan alan adını mutlaka incele ve şüpheli hiçbir QR sayfasına kart bilgisi girme!",
			StyleInformative: "Fiziksel ortamlardaki QR ve USB tuzaklarına karşı üç net güvenlik alışkanlığı! Bir: QR kodu okuttuğunda linke otomatik atlama; önizlemede görünen alan adının gerçekten ilgili kuruma ait olup olmadığını kontrol et. İki: Kafe, otopark veya sokak afişlerindeki QR etiketlerin üzerine sonradan başka bir sticker yapıştırılıp yapıştırılmadığına bak. Üç: Havalimanı ve AVM’lerdeki açık USB şarj istasyonlarına telefonunu doğrudan bağlama; veri aktarımı riskine karşı priz adaptörü veya kendi powerbankini tercih et. Küçük dikkat büyük kaybı önler!",
		},
		scenes: []SceneDefinition{
			{Title: "MASUM GÖRÜNEN QR KOD!", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 390, Y: 1040, W: 500}},
			{Title: "ÜZERİNE YAPIŞTIRILAN SAHTE ETİKET!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}, Arrow: true},
			{Title: "SAHTE ÖDEME SAYFASI AÇILDI", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}},
			{Title: "ÖNCE ÖNİZLEME ADRESİNİ OKU!", Color: "#E056FD", SFX: "paper", Prop: &SceneProp{Path: "@b1/idea-bulb-crayon.png", X: 390, Y: 1040, W: 500}},
			{Title: "RESMİ UYGULAMADAN İŞLEM YAP", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}},
			{Title: "GÖZÜNLE DEĞİL KONTROLLE GÜVEN!", Color: "#00F0FF", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 390, Y: 1040, W: 500}},
		},
	},
}

func normalizeMatchText(s string) string {
	r := strings.NewReplacer(
		"İ", "i", "I", "i", "ı", "i",
		"Ç", "c", "ç", "c",
		"Ğ", "g", "ğ", "g",
		"Ö", "o", "ö", "o",
		"Ş", "s", "ş", "s",
		"Ü", "u", "ü", "u",
	)
	return strings.ToLower(r.Replace(s))
}

func turkishTitle(s string) string {
	words := strings.Fields(strings.TrimSpace(s))
	if len(words) == 0 {
		return ""
	}
	for i, w := range words {
		runes := []rune(w)
		first := strings.ToUpperSpecial(unicode.TurkishCase, string(runes[0]))
		rest := ""
		if len(runes) > 1 {
			rest = strings.ToLowerSpecial(unicode.TurkishCase, string(runes[1:]))
		}
		words[i] = first + rest
	}
	return strings.Join(words, " ")
}

func generateViaTemplatesWithStyle(prompt string, style string) *ScriptResult {
	lowerRaw := strings.ToLower(prompt)
	lowerNorm := normalizeMatchText(prompt)

	for _, bp := range cyberTopicLibrary {
		for _, kw := range bp.keywords {
			if strings.Contains(lowerRaw, kw) || strings.Contains(lowerNorm, normalizeMatchText(kw)) {
				title := bp.titles[style]
				if title == "" {
					title = bp.titles[StyleHybrid]
				}
				narration := bp.narration[style]
				if narration == "" {
					narration = bp.narration[StyleHybrid]
				}
				scenes := make([]SceneDefinition, len(bp.scenes))
				copy(scenes, bp.scenes)
				return &ScriptResult{
					Title:     title,
					Slug:      slugify(title),
					Style:     style,
					Engine:    "smart-engine",
					Narration: narration,
					Scenes:    normalizeScenes(scenes),
				}
			}
		}
	}

	return buildDynamicScenario(prompt, style)
}

type dynamicDomainHint struct {
	keywords   []string
	hookScene  string
	attackHint string
	check1     string
	check2     string
	mythHint   string
	defense    string
	prop1      string
	prop2      string
}

func detectDomainHint(normPrompt string) dynamicDomainHint {
	hints := []dynamicDomainHint{
		{
			keywords:   []string{"llm", "prompt", "chatgpt", "yapay zeka", "ai", "ajan", "agent", "model"},
			hookScene:  "MASUM BİR KOMUT GİBİ GÖRÜNÜYOR…",
			attackHint: "görünmez bir yönlendirme metni ya da gizli veri sızdırma komutuyla modelin kuralları devre dışı bırakılmaya çalışılıyor.",
			check1:     "Dış kaynaklardan gelen metin, belge ve web içeriklerini asla doğrudan güvenilir sistem talimatı gibi işleme.",
			check2:     "Araç ve veri erişim yetkilerini yalnızca o an gereken en dar kapsamla sınırla.",
			mythHint:   "“Yapay zeka sadece sohbet ediyor” düşüncesi yanlıştır; yetkili ajanlara verilen her gizli talimat gerçek veri sızıntısına dönüşebilir.",
			defense:    "Hassas verileri izole et, kritik aksiyonlarda mutlaka insan onayı ve çıktı doğrulaması zorunlu tut.",
			prop1:      "@b1/cyber-terminal.png",
			prop2:      "@b1/cyber-database.png",
		},
		{
			keywords:   []string{"kamera", "iot", "akilli ev", "akıllı ev", "router", "modem", "bluetooth"},
			hookScene:  "EVİNDEKİ CİHAZ SESSİZCE AĞA BAĞLANDI…",
			attackHint: "kutudan çıktığı gibi bırakılan varsayılan şifreler ve güncellenmeyen yazılımlar yüzünden dışarıdan erişilebilir bir kapı bırakılıyor.",
			check1:     "Kurulum anında cihazın varsayılan yönetici şifresini mutlaka güçlü ve eşsiz bir parolayla değiştir.",
			check2:     "Akıllı cihazlarını ana bilgisayar ve telefonlarının bulunduğu ağdan ayırıp ayrı bir misafir ağına al.",
			mythHint:   "“Evimin modeminin arkasındayım, kimse göremez” yanılgısı tehlikelidir; tek bir açık port veya eski yazılım tüm ağı açabilir.",
			defense:    "Otomatik yazılım güncellemelerini aç, uzaktan erişimi yalnızca gerçekten gerektiğinde etkinleştir.",
			prop1:      "@b1/cyber-wifi.png",
			prop2:      "@b1/cyber-laptop.png",
		},
		{
			keywords:   []string{"e-devlet", "edevlet", "banka", "kart", "odeme", "ödeme", "iban", "vergi", "ceza", "icra"},
			hookScene:  "TELEFONUNA RESMİ GÖRÜNÜMLÜ BİR BİLDİRİM DÜŞTÜ…",
			attackHint: "resmi kurum ismi ve panik yaratan bir borç veya iade mesajıyla seni doğrudan sahte doğrulama sayfasına çekmeye çalışıyorlar.",
			check1:     "Gelen SMS veya e-postadaki bağlantıya dokunmak yerine tarayıcıdan ya da resmi mobil uygulamadan kendin giriş yap.",
			check2:     "Acele etmeni isteyen “Bugün son gün” veya “Hemen öde” ifadelerinin psikolojik baskı taktiği olduğunu unutma.",
			mythHint:   "Mesajda adının, logosunun veya resmi bir dilin yer alması o bildirimin gerçekten kurumdan geldiğini kanıtlamaz.",
			defense:    "Mobil bankacılık bildirimlerini ve iki aşamalı giriş onaylarını aktif tut, şüpheli durumlarda resmi hattı kendin ara.",
			prop1:      "@b1/cyber-laptop.png",
			prop2:      "@b1/cyber-padlock-unlocked.png",
		},
	}

	for _, h := range hints {
		for _, kw := range h.keywords {
			if strings.Contains(normPrompt, kw) {
				return h
			}
		}
	}

	return dynamicDomainHint{
		hookScene:  "HER ŞEY NORMAL GÖRÜNÜYOR… TA Kİ O DETAYA KADAR!",
		attackHint: "ilk bakışta masum görünen küçük bir açık ya da dikkatsiz bir onay adımı, arka planda tüm erişim zincirini tetikleyebiliyor.",
		check1:     "Kaynağı ve göndereni doğrulamadan hiçbir bağlantıyı, dosyayı veya yetki isteğini varsayılan olarak kabul etme.",
		check2:     "Seni hızlanmaya zorlayan “hemen onayla” ya da “tek tıkla geç” yönlendirmelerinde mutlaka birkaç saniye dur ve kontrol et.",
		mythHint:   "Görünüşte tanıdık bir arayüz veya kilit simgesi görmek tek başına güvenliğin sağlandığı anlamına gelmez.",
		defense:    "En düşük yetki ilkesini uygula, yazılımlarını güncel tut ve tüm kritik hesaplarda iki aşamalı doğrulamayı etkinleştir.",
		prop1:      "@b1/cyber-laptop.png",
		prop2:      "@b1/cyber-terminal.png",
	}
}

func compactSceneTopic(cleanSubject string) string {
	words := strings.Fields(cleanSubject)
	if len(words) > 3 {
		words = words[:3]
	}
	short := strings.ToUpperSpecial(unicode.TurkishCase, strings.Join(words, " "))
	if short == "" {
		return "SİBER TEHDİT"
	}
	return short
}

func buildDynamicScenario(prompt, style string) *ScriptResult {
	cleanSubject := strings.TrimSpace(prompt)
	prettySubject := turkishTitle(cleanSubject)
	normPrompt := normalizeMatchText(cleanSubject)
	hint := detectDomainHint(normPrompt)
	shortUpper := compactSceneTopic(cleanSubject)

	var title, narration string
	switch style {
	case StyleCinematic:
		title = fmt.Sprintf("%s: Görünmeyen Tuzak Nasıl Başlar?", prettySubject)
		if len([]rune(title)) > 52 {
			title = fmt.Sprintf("%s Gerçeği!", prettySubject)
		}
		narration = fmt.Sprintf(
			"Ekranına sıradan bir bildirim düşüyor… %s İlk saniyede her şey kontrolün altındaymış gibi hissediyorsun, değil mi? Oysa konu %s olduğunda saldırganın en büyük kozu senin bu rahat anın. Arka planda %s Ve çoğu kişi bunu ancak hesaplar kilitlendiğinde ya da veriler dışarı aktığında fark ediyor! Unutma: %s Bu senaryoyu yaşamanın önüne geçmek için tek bir kural yeter: hızını kes ve kontrol et. %s Ardından %s Bir anlık dikkatsizliğin bedelini aylarca ödeme. Dur, düşün, doğrula! Takipte kal, güvende kal!",
			hint.hookScene, cleanSubject, hint.attackHint, hint.mythHint, hint.check1, hint.defense,
		)
	case StyleInformative:
		title = fmt.Sprintf("%s Hakkında 3 Kritik Güvenlik Kuralı", prettySubject)
		if len([]rune(title)) > 52 {
			title = fmt.Sprintf("%s Korunma Rehberi", prettySubject)
		}
		narration = fmt.Sprintf(
			"%s konusunda farkında olmadan yaptığın hataları beş saniyede tespit edelim! Saldırıların önemli kısmı karmaşık kodlardan değil, küçük bir kontrol eksikliğinden doğuyor. Çünkü %s İşte hemen uygulaman gereken üç net kural: Bir, %s İki, %s Üç, %s Unutma: %s Bu kısa rehberi kaydet ve ekip arkadaşlarınla paylaş. Takipte kal, güvende kal!",
			prettySubject, hint.attackHint, hint.check1, hint.check2, hint.defense, hint.mythHint,
		)
	default: // StyleHybrid
		title = fmt.Sprintf("%s Konusunda Bu Hataya Düşme!", prettySubject)
		if len([]rune(title)) > 52 {
			title = fmt.Sprintf("%s ve Gizli Riskler", prettySubject)
		}
		narration = fmt.Sprintf(
			"%s Peki konu %s olduğunda seni asıl neyin beklediğini hiç düşündün mü? Her şey normal akışında ilerlerken %s Çoğu kullanıcı “Bana bir şey olmaz” diye düşünür; oysa %s Kendini saniyeler içinde korumanın yolu ise çok net: İlk olarak, %s İkinci olarak, %s Son olarak, %s Çünkü dijital güvenliğini korumanın sırrı korkmak değil, saldırı gelmeden bir adım önde durmaktır! Takipte kal, güvende kal!",
			hint.hookScene, cleanSubject, hint.attackHint, hint.mythHint, hint.check1, hint.check2, hint.defense,
		)
	}

	scenes := []SceneDefinition{
		{Title: shortUpper + " RİSKİ!", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: hint.prop1, X: 390, Y: 1040, W: 500}},
		{Title: "GÖRÜNMEYEN AÇIK NOKTA!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}, Arrow: true},
		{Title: "ARKA PLANDA SESSİZ SIZMA", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: hint.prop2, X: 390, Y: 1040, W: 500}},
		{Title: "YANILGIYI FARK ET!", Color: "#E056FD", SFX: "paper", Prop: &SceneProp{Path: "@b1/idea-bulb-crayon.png", X: 390, Y: 1040, W: 500}},
		{Title: "DUR. DÜŞÜN. DOĞRULA!", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}},
		{Title: "KONTROLÜ ELİNDE TUT!", Color: "#00F0FF", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 390, Y: 1040, W: 500}},
	}

	return &ScriptResult{
		Title:     title,
		Slug:      slugify(title),
		Style:     style,
		Engine:    "smart-engine",
		Narration: narration,
		Scenes:    normalizeScenes(scenes),
	}
}

// ViralMetadata holds A/B/C test titles, an engaging social media post caption, and viral hashtags.
type ViralMetadata struct {
	TitleA      string   `json:"title_a"`
	TitleB      string   `json:"title_b"`
	TitleC      string   `json:"title_c"`
	Description string   `json:"description"`
	Hashtags    []string `json:"hashtags"`
}

// GenerateViralMetadata produces high-converting A/B/C titles, caption text, and hashtags.
func GenerateViralMetadata(title, narration string) ViralMetadata {
	if !testing.Testing() && os.Getenv("TUBIE_OFFLINE_AI") != "1" {
		if vm, err := generateViralViaOllama(title, narration); err == nil && vm != nil {
			return *vm
		}
	}
	return generateViralViaTemplates(title, narration)
}

func generateViralViaOllama(title, narration string) (*ViralMetadata, error) {
	pingClient := &http.Client{Timeout: 500 * time.Millisecond}
	pingResp, err := pingClient.Get(defaultOllamaBaseURL + "/api/tags")
	if err != nil {
		return nil, err
	}
	defer pingResp.Body.Close()
	if pingResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama not ready")
	}

	var tags ollamaTagsResponse
	_ = json.NewDecoder(pingResp.Body).Decode(&tags)
	modelName := selectOllamaModel(tags.Models)

	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	sysPrompt := fmt.Sprintf(`Sen Instagram Reels ve TikTok viral içerik uzmanısın.
Video Başlığı: %q
Seslendirme Metni: %q

Bu video için Instagram paylaşım paketi hazırla.
1. A/B/C testi için 3 farklı dikkat çekici başlık (Merak, Uyarı, Rehber).
2. İlgi çekici, emoji dolu, 150-200 kelimelik viral Instagram Reels açıklaması.
3. 15-20 adet popüler ve konuyla ilgili hashtag.
Instagram 2200 karakter sınırına kesinlikle uy.

Yanıtını SADECE geçerli bir JSON olarak ver:
{
  "title_a": "🚨 Başlık A",
  "title_b": "⚠️ Başlık B",
  "title_c": "🛡️ Başlık C",
  "description": "Açıklama metni...",
  "hashtags": ["#sibergüvenlik", "#reels", "#keşfet"]
}`, title, narration)

	reqBody := map[string]any{
		"model":  modelName,
		"prompt": sysPrompt,
		"stream": false,
		"format": "json",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, defaultOllamaBaseURL+"/api/generate", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status %d", resp.StatusCode)
	}

	var ollamaResp struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, err
	}

	var vm ViralMetadata
	if err := json.Unmarshal([]byte(stripJSONFences(ollamaResp.Response)), &vm); err != nil {
		return nil, err
	}
	if vm.TitleA == "" || vm.Description == "" || len(vm.Hashtags) == 0 {
		return nil, fmt.Errorf("incomplete viral response")
	}
	return &vm, nil
}

// FixedYouTubeTags contains the mandatory, high-converting YouTube tags for cybersecurity videos.
const FixedYouTubeTags = "hacker,hack,hackers,hacks,atıl samancıoğlu,can değer,mdisec,para kazanma,etik hacker olma kursu,etik hacker,panel,cc,active directory,redteam,blueteam,pentest,penetration test,analist,soc analisti,soc,ioc,iot,bugbounty,bug bount,etik hacker olma,hacker olma kursu,siber tehdit,siber güvenlik,cyber security,cyber sec,bilgi güvenliği,mit,btk,yazılım,bilişim,bilgisayar,network,veri güvenliği,bilişim güvenliği,siyah şapkalı hacker,hackerlık,darkweb,deepweb,osint"

func generateViralViaTemplates(title, narration string) ViralMetadata {
	lower := strings.ToLower(title + " " + narration)

	if strings.Contains(lower, "sahte link") || strings.Contains(lower, "oltalama") || strings.Contains(lower, "phishing") || strings.Contains(lower, "bu linke") {
		return ViralMetadata{
			TitleA:      "🚨 Bu Linke Tıklarsan Ne Olur? (5 Saniyede Sahte Linki Yakala!)",
			TitleB:      "⚠️ 'Hesabınızı Hemen Doğrulayın' Mesajı Aldınız mı? Sakın Tıklamayın!",
			TitleC:      "🛡️ Sahte Linkleri ve Oltalama Tuzaklarını Anlamanın 4 Net Yolu!",
			Description: "Tanıdık bir logo, inandırıcı bir mesaj ve 'Hemen doğrulayın!' baskısı… İşte bütün oltalama (phishing) saldırıları tam olarak böyle başlıyor! 🚨🔗\n\nSaldırganların bütün dijital hayatınıza ulaşmak için ihtiyaç duyduğu tek şey sizin bir anlık dikkatsizliğinizdir.\n\n⚠️ SAHTE LİNKİ 5 SANİYEDE NASIL ANLARSIN?\n1️⃣ Alan Adını İncele: Küçük harf farkları (l yerine I) ve garip uzantılar tuzaktır.\n2️⃣ Göndereni Kontrol Et: Tanıdık kurum adı yazması mesajın gerçekten oradan geldiğini kanıtlamaz.\n3️⃣ Acele Ettirme Manipülasyonu: 'Hesabınız kapatılacak!' baskısına kanıp panikle tıklamayın.\n4️⃣ Kilit Simgesi Yanılgısı: HTTPS ve kilit simgesi sitenin güvenli olduğunu değil, sadece bağlantının şifreli olduğunu gösterir!\n\n🛡️ TIKLAMADAN ÖNCE DUR. DÜŞÜN. DOĞRULA!\n✅ Bağlantıya değil, doğrudan resmi uygulamaya veya siteye kendiniz girin.\n✅ Güçlü parolalar ve Authenticator tabanlı 2FA kullanın.\n\n👇 Bu videoyu kaydet ve 'her şeye tıklayan' arkadaşlarına hemen gönder!",
			Hashtags: []string{
				"#sibergüvenlik", "#oltalama", "#phishing", "#sahtelink", "#hacker",
				"#dijitalgüvenlik", "#bilgi", "#teknoloji", "#reels", "#keşfet",
				"#dolandırıcılık", "#2fa", "#siberfarkındalık", "#etikahacker",
				"#internetgüvenliği", "#trendler", "#fyp", "#viral",
			},
		}
	} else if strings.Contains(lower, "instagram") || strings.Contains(lower, "sosyal medya") || strings.Contains(lower, "hesap çal") {
		return ViralMetadata{
			TitleA:      "🚨 Instagram Hesabınız 1 Saniyede Nasıl Çalınır? (Sakın Bu Hatayı Yapmayın!)",
			TitleB:      "⚠️ 'Telif Hakkı İhlali' Mesajı Aldınız mı? Sakın Tıklamayın!",
			TitleC:      "🛡️ Instagram Hesabınızı Hackerlardan Korumanın 3 Altın Kuralı!",
			Description: "Instagram hesabınız tek bir sahte bağlantıyla saniyeler içinde tamamen elinizden gidebilir! 😱\n\nHackerlar artık karmaşık şifre kırmakla uğraşmıyor. Kullanıcı psikolojisini hedef alan sahte DM ve oltalama (phishing) tuzakları kuruyorlar.\n\n⚠️ EN SIK KULLANILAN TUZAKLAR:\n1️⃣ Telif İhlali Bildirimi: Profilinizin kapatılacağını iddia eden sahte DM'ler.\n2️⃣ Mavi Tik / Rozet Bildirimi: Sahte formla şifre toplayan profiller.\n3️⃣ Şüpheli Giriş Uyarısı: Panikle şifre girmenizi isteyen sahte linkler.\n\n🛡️ HESABINIZI KORUMANIN 4 ALTIN KURALI:\n✅ 1. Instagram yetkilileri ASLA DM ile şifre veya link göndermez.\n✅ 2. Bildirimleri sadece Ayarlar > Güvenlik > Instagram E-postaları sekmesinden kontrol edin.\n✅ 3. SMS yerine mutlaka Google Authenticator gibi bir 2FA uygulaması açın.\n✅ 4. Şüpheli linke tıkladıysanız derhal şifrenizi yenileyin ve oturumları kapatın.\n\n👇 Bu tür sahte mesajlarla daha önce karşılaştınız mı? Yorumlarda yazın!\n👥 Videoyu kaydedin ve arkadaşlarınızı uyarın, kimsenin hesabı çalınmasın!",
			Hashtags: []string{
				"#sibergüvenlik", "#instagramgüvenliği", "#hesapçalma", "#hacker",
				"#sosyalmedya", "#güvenlik", "#teknoloji", "#bilgi", "#reels", "#keşfet",
				"#siberbilgi", "#yazılım", "#dijitalgüvenlik", "#trendler", "#fyp",
				"#şifregüvenliği", "#tuzak", "#etikahacker", "#siberfarkındalık",
				"#phishing", "#oltalama", "#instagramreels", "#bilgisayar", "#veri",
				"#sibertehdit", "#bilişim", "#viral",
			},
		}
	} else if strings.Contains(lower, "wifi") || strings.Contains(lower, "wi-fi") || strings.Contains(lower, "kafe") || strings.Contains(lower, "ağ") {
		return ViralMetadata{
			TitleA:      "🚨 Kafe Wi-Fi Ağlarına Bağlananlar DİKKAT! (Bunu Biliyor muydunuz?)",
			TitleB:      "⚠️ Şifrenizi 10 Saniyede Havada Yakalayabilirler!",
			TitleC:      "🛡️ Halka Açık Wi-Fi Ağlarında Güvende Kalmanın Tek Yolu!",
			Description: "Gittiğiniz kafede ücretsiz Wi-Fi ağına bağlandığınız an tüm kişisel verileriniz havada izleniyor olabilir! ☕💻\n\nHalka açık, şifresiz kablosuz ağlar siber saldırganlar için en verimli av sahalarıdır. Hackerlar aynı ağa bağlandıktan sonra trafiği kolayca dinleyebilir.\n\n⚠️ KAFELERDEKİ GÖRÜNMEZ TEHLİKELER:\n1️⃣ Evil Twin (Sahte İkiz Ağ): Kafenin adıyla birebir sahte bir Wi-Fi noktası açarlar.\n2️⃣ Deauth Saldırısı: Sizi ağdan düşürürler; tekrar bağlanırken şifreli paketler havada yakalanır.\n3️⃣ Man-in-the-Middle: Girdiğiniz siteler ile cihazınız arasına girip şifreleri izlerler.\n\n🛡️ AÇIK AĞLARDA GÜVENDE KALMANIN ALTIN KURALLARI:\n✅ 1. Ortak Wi-Fi ağlarında asla bankacılık veya hassas işlemler yapmayın.\n✅ 2. Cihazınızda mutlaka güvenilir ve şifreli bir VPN servisi açık olsun.\n✅ 3. Telefonunuzun 'Bilinen ağlara otomatik bağlan' özelliğini kapatın.\n✅ 4. Girdiğiniz web sitelerinde HTTPS kilit simgesini kontrol edin.\n\n👇 Dışarıda kafelerin Wi-Fi ağlarına bağlanıyor musunuz? Yorumlarda buluşalım!\n👥 Videoyu kaydedin ve sevdiklerinizle paylaşın!",
			Hashtags: []string{
				"#sibergüvenlik", "#wifi", "#ağgüvenliği", "#vpn", "#hacker",
				"#teknoloji", "#reels", "#keşfet", "#bilgi", "#dijitaldünya",
				"#internet", "#güvenlik", "#teknolojihaberleri", "#siberdünya",
				"#deauth", "#network", "#etikahacker", "#veri", "#bilgisayar",
				"#bilişim", "#sibertehdit", "#trendler", "#fyp", "#viral",
			},
		}
	} else if strings.Contains(lower, "şifre") || strings.Contains(lower, "parola") || strings.Contains(lower, "brute force") {
		return ViralMetadata{
			TitleA:      "🚨 Kullandığınız Şifre Kaç Saniyede Kırılır? (Şok Eden Gerçek!)",
			TitleB:      "⚠️ '123456' veya Doğum Tarihi Kullananlar Hemen Değiştirsin!",
			TitleC:      "🛡️ Kırılması Yüzyıllar Sürecek Güçlü Şifre Oluşturma Formülü!",
			Description: "Kullandığınız şifre gerçekten güvende mi yoksa hackerların kırması sadece birkaç saniye mi sürer? ⏱️🔐\n\nHer gün milyonlarca kullanıcı basit parolalar kullanıyor. Oysa yapay zeka ve brute force araçları saniyede milyarlarca kombinasyon deneyebiliyor!\n\n⚠️ EN BÜYÜK ŞİFRE HATALARI:\n1️⃣ Aynı Parolayı Her Yerde Kullanmak: Bir site sızdığında tüm hesaplarınız ele geçirilir.\n2️⃣ Kişisel Bilgileri Şifre Yapmak: İsim, memleket ve tarihler saniyeler içinde çözülür.\n3️⃣ Kısa Şifreler: 8 karakterli küçük harfli şifre ortalama 5 saniyede kırılır!\n\n🛡️ KIRILMASI YÜZYILLAR SÜRECEK GÜÇLÜ PAROLA FORMÜLÜ:\n✅ 1. En az 14-16 karakter uzunluğunda olmalıdır.\n✅ 2. Büyük harf, küçük harf, rakam ve özel semboller içermelidir.\n✅ 3. Her platform için mutlaka farklı bir şifre kullanın.\n✅ 4. Parolalarınızı bir Parola Yöneticisinde güvenle saklayın.\n\n👇 Şifreniz ne kadar güçlü? Yorumlarda konuşalım!\n👥 Sevdiklerinizi uyarmak için videoyu hemen kaydedip gönderin!",
			Hashtags: []string{
				"#parola", "#şifregüvenliği", "#bruteforce", "#sibergüvenlik",
				"#teknoloji", "#güvenlik", "#bilgi", "#reels", "#keşfet",
				"#dijitalyaşam", "#hacker", "#şifre", "#bilgisayar", "#etikahacker",
				"#passwordmanager", "#verigüvenliği", "#siberfarkındalık",
				"#trendler", "#fyp", "#viral",
			},
		}
	} else if strings.Contains(lower, "kripto") || strings.Contains(lower, "bitcoin") || strings.Contains(lower, "cüzdan") {
		return ViralMetadata{
			TitleA:      "🚨 Kripto Cüzdanınız 1 Tıkla Nasıl Boşaltılır? (Airdrop Tuzağı!)",
			TitleB:      "⚠️ 'Bedava Coin Kazandınız' Mesajlarına Sakın İnanmayın!",
			TitleC:      "🛡️ Kripto Paralarınızı Güvende Tutmanın 3 Temel Kuralı!",
			Description: "Kripto cüzdanınızdaki tüm birikim tek bir sahte onay ile bir saniyede sıfırlanabilir! 💸⚠️\n\nKripto dünyasında hackerlar sahte airdrop siteleri ve Telegram gruplarıyla kullanıcıları tuzağa düşürüyor!\n\n⚠️ EN TEHLİKELİ KRİPTO TUZAKLARI:\n1️⃣ Akıllı Sözleşme İzni (Approve Drainer): Cüzdanı bağlayıp onay verdiğiniz an tokenlarınız transfer edilir.\n2️⃣ Sahte Cüzdan Güncellemesi: DM ile gelen ve güncelleme isteyen sahte linkler.\n3️⃣ Sahte Destek Ekipleri: Size özelden yazıp yardım teklif eden sahte yöneticiler.\n\n🛡️ KRİPTO VARLIKLARINIZI KORUMANIN 4 KURALI:\n✅ 1. Gizli Kurtarma Kelimelerinizi (Seed Phrase) ASLA hiçbir web sitesine yazmayın!\n✅ 2. Hiç kimse durduk yere karşılıksız bedava para dağıtmaz.\n✅ 3. Bilmediğiniz sitelere cüzdan bağlamayın; onayları düzenli iptal edin.\n✅ 4. Yüksek birikimleriniz için mutlaka Donanım Cüzdanı (Cold Wallet) kullanın.\n\n👇 Kripto yatırımı olan arkadaşını hemen etiketle!\n👥 Gönderiyi kaydet, paranı koru!",
			Hashtags: []string{
				"#kripto", "#bitcoin", "#ethereum", "#kriptopara", "#kriptogüvenlik",
				"#airdrop", "#cüzdan", "#sibergüvenlik", "#teknoloji", "#reels",
				"#keşfet", "#finans", "#blokzincir", "#bilgi", "#donanımcüzdanı",
				"#dolandırıcılık", "#coldwallet", "#btc", "#etikahacker", "#fyp", "#viral",
			},
		}
	} else if strings.Contains(lower, "whatsapp") || strings.Contains(lower, "sms") || strings.Contains(lower, "telefon") {
		return ViralMetadata{
			TitleA:      "🚨 Arkadaşınızdan 'Kodu Bana At' Mesajı Gelirse Sakın Atmayın!",
			TitleB:      "⚠️ WhatsApp Hesabınızı Tek Bir SMS Koduyla Çalıyorlar!",
			TitleC:      "🛡️ WhatsApp'ta İki Adımlı Doğrulama Nasıl Açılır?",
			Description: "WhatsApp hesabınız aniden kapanabilir ve rehberinizdeki herkese adınıza borç isteme mesajları gidebilir! 📱🚨\n\nSaldırganlar yakın arkadaşlarınızın çalınmış profilleri üzerinden SMS onay kodu hırsızlığı yapıyor.\n\n⚠️ DOLANDIRICILAR NASIL ÇALIŞIYOR?\n1️⃣ Saldırgan yeni bir cihaza telefon numaranızı yazar.\n2️⃣ WhatsApp telefonunuza 6 haneli resmi bir SMS onay kodu yollar.\n3️⃣ Arkadaşınızın profilinden mesaj gelir: 'Kodu yanlışlıkla sana yolladım, bana atar mısın?'\n4️⃣ O kodu verdiğiniz anda hesabınız saldırganın eline geçer!\n\n🛡️ HESABINIZI GERİ DÖNÜLMEZ ŞEKİLDE KORUYUN:\n✅ 1. SMS onay kodunu EN YAKININIZA BİLE asla vermeyin!\n✅ 2. Arkadaşınızdan gelen garip taleplerde kendisini sesli arayıp teyit edin.\n✅ 3. WhatsApp Ayarları > Hesap > İki Adımlı Doğrulama'yı hemen aktif edin.\n✅ 4. Hesabınıza e-posta adresi bağlayarak güvenliği ikiye katlayın.\n\n👇 Bu SMS onay hilesiyle daha önce karşılaştınız mı? Yorumlarda paylaşın!\n👥 Sevdiklerinize gönderin, kimse mağdur olmasın!",
			Hashtags: []string{
				"#whatsapp", "#dolandırıcılık", "#sibergüvenlik", "#smstuzağı",
				"#teknoloji", "#güvenlik", "#reels", "#keşfet", "#haberler",
				"#bilgi", "#iletişim", "#dikkat", "#siberbilgi", "#etikahacker",
				"#dijitalgüvenlik", "#telefon", "#trendler", "#fyp", "#viral",
			},
		}
	}

	cleanTitle := strings.TrimSpace(title)
	if cleanTitle == "" {
		cleanTitle = "Siber Güvenlik Gerçeği"
	}
	return ViralMetadata{
		TitleA:      fmt.Sprintf("🚨 %s: Tek Bir Tıkla Her Şey Nasıl Risk Altına Girer?", cleanTitle),
		TitleB:      fmt.Sprintf("⚠️ %s Hakkında Bilmeden Yaptığınız Kritik Hata!", cleanTitle),
		TitleC:      fmt.Sprintf("🛡️ %s Konusunda Kendinizi Korumanın 3 Net Kuralı!", cleanTitle),
		Description: fmt.Sprintf("%s hakkında herkesin bilmesi gereken siber güvenlik gerçekleri! 🚨\n\nSaldırılar çoğu zaman karmaşık açıklarla değil, aciliyet hissi ve bir anlık dikkatsizlikle başlar.\n\n⚠️ DİJİTAL DÜNYADA DİKKAT EDİLMESİ GEREKENLER:\n1️⃣ Şüpheli bağlantıları ve gönderici adreslerini mutlaka kontrol edin.\n2️⃣ Kaynağından emin olmadığınız dosya ve yetki onaylarından uzak durun.\n3️⃣ Kritik hesaplarınızda mutlaka Authenticator tabanlı 2FA kullanın.\n\n🛡️ KÜÇÜK BİR KONTROL BÜYÜK BİR KAYBI ÖNLER:\n✅ Dur, düşün ve doğrula kuralını alışkanlık haline getirin.\n✅ Güçlü ve her platformda farklı parolalar tercih edin.\n\n👇 Siz bu konuda ne düşünüyorsunuz? Yorumlarda buluşalım!\n👥 Gönderiyi kaydedin ve sevdiklerinizle paylaşın!", cleanTitle),
		Hashtags: []string{
			"#sibergüvenlik", "#teknoloji", "#güvenlik", "#hacker",
			"#bilgi", "#dijitaldünya", "#reels", "#keşfet", "#trendler",
			"#yazılım", "#bilimveteknoloji", "#önlem", "#farkındalık",
			"#siberbilgi", "#etikahacker", "#darkweb", "#pentest", "#fyp", "#viral",
		},
	}
}

// FormatViralText creates the aciklama.txt document with A/B/C titles, caption, hashtags, and YouTube tags.
func FormatViralText(title, narration string, vm ViralMetadata) string {
	var sb strings.Builder
	sb.WriteString("🎬 TUBIE VİRAL PAYLAŞIM PAKETİ (INSTAGRAM / REELS / TIKTOK / SHORTS)\n")
	sb.WriteString("====================================================================\n\n")

	sb.WriteString("📌 A / B / C TESTİ İÇİN 3 FARKLI BAŞLIK:\n")
	sb.WriteString("(En yüksek tıklanma ve izlenme oranını yakalamak için test edin)\n")
	sb.WriteString("--------------------------------------------------------------------\n")
	sb.WriteString(fmt.Sprintf("[BAŞLIK A - MERAK / ŞOK]:\n%s\n\n", vm.TitleA))
	sb.WriteString(fmt.Sprintf("[BAŞLIK B - UYARI / FOMO]:\n%s\n\n", vm.TitleB))
	sb.WriteString(fmt.Sprintf("[BAŞLIK C - REHBER / ÇÖZÜM]:\n%s\n\n", vm.TitleC))

	sb.WriteString("📝 VİRAL REELS & POST AÇIKLAMASI (Kopyalayıp Yapıştırın):\n")
	sb.WriteString("--------------------------------------------------------------------\n")
	sb.WriteString(vm.Description + "\n\n")

	sb.WriteString("🚀 VİRAL KEŞFET HASHTAGLERİ:\n")
	sb.WriteString("--------------------------------------------------------------------\n")
	sb.WriteString(strings.Join(vm.Hashtags, " ") + "\n\n")

	sb.WriteString("🏷️ YOUTUBE TAG (Video Yüklerken Etiket Bölümüne Yapıştırın):\n")
	sb.WriteString("--------------------------------------------------------------------\n")
	sb.WriteString(FixedYouTubeTags + "\n")

	return sb.String()
}

func slugify(text string) string {
	lower := strings.ToLower(text)
	r := strings.NewReplacer(
		"ç", "c", "ğ", "g", "ı", "i", "ö", "o", "ş", "s", "ü", "u",
		"Ç", "c", "Ğ", "g", "İ", "i", "Ö", "o", "Ş", "s", "Ü", "u",
	)
	normalized := r.Replace(lower)
	re := regexp.MustCompile(`[^a-z0-9]+`)
	slug := re.ReplaceAllString(normalized, "_")
	slug = strings.Trim(slug, "_")
	if len(slug) > 40 {
		slug = slug[:40]
	}
	if slug == "" {
		slug = "siber_video"
	}
	return slug
}
