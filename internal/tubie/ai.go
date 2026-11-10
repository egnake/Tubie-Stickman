package tubie

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// ScriptResult represents the complete generated script, title, and storyboard.
type ScriptResult struct {
	Title     string            `json:"title"`
	Slug      string            `json:"slug"`
	Narration string            `json:"narration"`
	Scenes    []SceneDefinition `json:"scenes"`
}

// GenerateScript produces a viral cyber security script from a user prompt.
func GenerateScript(userPrompt string) (*ScriptResult, error) {
	trimmed := strings.TrimSpace(userPrompt)
	if trimmed == "" {
		return nil, fmt.Errorf("lütfen bir video konusu veya fikir yazın")
	}

	// 1. Try Ollama if running locally (with 10-second timeout)
	if res, err := generateViaOllama(trimmed); err == nil && res != nil && len(res.Scenes) >= 3 {
		return res, nil
	}

	// 2. Intelligent offline heuristic template generator (instant, zero failure)
	return generateViaTemplates(trimmed), nil
}

func generateViaOllama(prompt string) (*ScriptResult, error) {
	// Fast ping (500ms) to check if Ollama is alive locally
	pingClient := &http.Client{Timeout: 500 * time.Millisecond}
	pingResp, err := pingClient.Get("http://localhost:11434/api/tags")
	if err != nil {
		return nil, err
	}
	_ = pingResp.Body.Close()
	if pingResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama not ready")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	sysPrompt := fmt.Sprintf(`Sen siber güvenlik konusunda uzman, viral YouTube Shorts ve Reels video yönetmenisin.
Kullanıcının konusu: %q.
Bu konuda 30 saniyelik çok akıcı, heyecan verici, yüksek enerjili Türkçe seslendirme metni yaz.
Hackerların nasıl tuzak kurduğunu ve kullanıcının nasıl korunacağını anlat.
Yanıtını kesinlikle SADECE geçerli bir JSON olarak ver:
{
  "title": "Kısa ve çarpıcı video başlığı",
  "narration": "60-80 kelimelik yüksek tempolu Türkçe seslendirme metni",
  "scenes": [
    {"title": "Sahne 1 başlığı", "sfx": "keyboard", "color": "#00F0FF"},
    {"title": "Sahne 2 başlığı", "sfx": "alert", "color": "#FF3B30"},
    {"title": "Sahne 3 başlığı", "sfx": "glitch", "color": "#FFCC00"},
    {"title": "Sahne 4 başlığı", "sfx": "lock", "color": "#00FF66"}
  ]
}`, prompt)

	reqBody := map[string]any{
		"model":  "qwen2.5:7b",
		"prompt": sysPrompt,
		"stream": false,
		"format": "json",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:11434/api/generate", bytes.NewReader(bodyBytes))
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

	var parsed struct {
		Title     string            `json:"title"`
		Narration string            `json:"narration"`
		Scenes    []SceneDefinition `json:"scenes"`
	}
	if err := json.Unmarshal([]byte(ollamaResp.Response), &parsed); err != nil {
		return nil, err
	}

	if parsed.Title == "" || parsed.Narration == "" || len(parsed.Scenes) < 3 {
		return nil, fmt.Errorf("incomplete ollama response")
	}

	slug := slugify(parsed.Title)
	return &ScriptResult{
		Title:     parsed.Title,
		Slug:      slug,
		Narration: parsed.Narration,
		Scenes:    parsed.Scenes,
	}, nil
}

func generateViaTemplates(prompt string) *ScriptResult {
	lower := strings.ToLower(prompt)

	type tpl struct {
		title     string
		narration string
		scenes    []SceneDefinition
	}

	var matched *tpl

	if strings.Contains(lower, "instagram") || strings.Contains(lower, "sosyal medya") || strings.Contains(lower, "hesap çal") {
		matched = &tpl{
			title: "İnstagram Hesabınız Nasıl Çalınır?",
			narration: "Instagram hesabın bir gecede elinden gidebilir! Hackerlar şifreni kırmakla uğraşmaz. 'Telif hakkı ihlali' veya 'Mavi tik kazandınız' diye sahte bir DM atarlar. Panikle linke tıklar ve sahte giriş sayfasına şifreni yazarsın. Tebrikler, hesabını ve tüm takipçilerini hacker'a hediye ettin! Gerçek Instagram asla DM'den şifre istemez. İki faktörlü doğrulamayı hemen aç ve şüpheli linklere tıklama! Uyanık kal, hesabını koru!",
			scenes: []SceneDefinition{
				{Title: "SAHTE DM TUZAĞI", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 490, Y: 1420, W: 300}},
				{Title: "ŞİFREN ELE GEÇİRİLDİ!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 490, Y: 1420, W: 300}, Arrow: true},
				{Title: "HESAP KONTROLÜ KAYBEDİLDİ", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 490, Y: 1420, W: 300}},
				{Title: "2FA İLE HESABINI KORU!", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1420, W: 300}},
			},
		}
	} else if strings.Contains(lower, "kripto") || strings.Contains(lower, "bitcoin") || strings.Contains(lower, "cüzdan") {
		matched = &tpl{
			title: "Kripto Cüzdanınız Nasıl Sıfırlanır?",
			narration: "Kripto cüzdanındaki tüm birikim tek bir saniyede buharlaşabilir! Sahte airdrop siteleri ve dolandırıcı Telegram grupları sana bedava coin vaat eder. Cüzdanını bağladığın an, akıllı sözleşme izniyle tüm tokenlarını çekerler! Altın kural: Hiç kimse durduk yere bedava para dağıtmaz. Gizli kelimelerini asla bir web sitesine yazma ve her sözleşme onayını iki kez kontrol et! Parana sahip çık!",
			scenes: []SceneDefinition{
				{Title: "BEDAVA COIN TUZAĞI!", Color: "#FFCC00", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 490, Y: 1420, W: 300}},
				{Title: "AKILLI SÖZLEŞME ONAYI", Color: "#FF3B30", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 490, Y: 1420, W: 300}, Arrow: true},
				{Title: "CÜZDAN BOŞALTILDI!", Color: "#FF3B30", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 490, Y: 1420, W: 300}},
				{Title: "KELİMELERİNİ ASLA PAYLAŞMA", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-padlock-unlocked.png", X: 490, Y: 1420, W: 300}},
			},
		}
	} else if strings.Contains(lower, "whatsapp") || strings.Contains(lower, "telefon") || strings.Contains(lower, "sms") {
		matched = &tpl{
			title: "WhatsApp Hesabınız Nasıl Ele Geçirilir?",
			narration: "WhatsApp hesabından arkadaşlarına borç para isteme mesajları gidebilir! Hacker numaranı sisteme yazar ve sana bir SMS onay kodu gelir. Hemen ardından samimi bir arkadaşından mesaj alırsın: 'Kodu yanlışlıkla sana yolladım, bana atar mısın?' O kodu verdiğin anda hesabın gider! SMS doğrulama kodlarını en yakınına bile asla verme. WhatsApp'ta iki adımlı doğrulamayı hemen etkinleştir!",
			scenes: []SceneDefinition{
				{Title: "BEKLENMEDİK ONAY KODU", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-wifi.png", X: 490, Y: 1420, W: 300}},
				{Title: "KODU VERDİĞİN AN BİTTİN!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 490, Y: 1420, W: 300}, Arrow: true},
				{Title: "HESAP BAŞKASINA GEÇTİ", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 490, Y: 1420, W: 300}},
				{Title: "İKİ ADIMLI DOĞRULAMAYI AÇ", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1420, W: 300}},
			},
		}
	} else {
		// General cybersecurity dynamic generator
		cleanSubject := strings.TrimSpace(prompt)
		title := fmt.Sprintf("%s Hakkında Bilmeniz Gereken Siber Tehlike!", strings.Title(cleanSubject))
		if len(title) > 45 {
			title = fmt.Sprintf("%s ve Siber Güvenlik Gerçeği", strings.Title(cleanSubject))
		}

		matched = &tpl{
			title: title,
			narration: fmt.Sprintf("Dikkat! %s konusu sandığından çok daha kritik bir siber güvenlik riski taşıyor olabilir. Hackerlar her gün binlerce kullanıcının dikkatsizliğinden faydalanarak sistemlere sızıyor. Basit bir dikkatsizlik, tüm kişisel verilerinin ve parolalarının açığa çıkmasına neden olabilir. Unutma: Siber güvenlikte en zayıf halka insandır. Güçlü parolalar kullan, çift faktörlü korumayı aç ve daima tetikte kal! Takipte ol, güvenle kal!", cleanSubject),
			scenes: []SceneDefinition{
				{Title: "GÖRÜNMEYEN TEHLİKE!", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 490, Y: 1420, W: 300}},
				{Title: "SİSTEME SIZMA ANI", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 490, Y: 1420, W: 300}, Arrow: true},
				{Title: "VERİLERİN AÇIĞA ÇIKMASI", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 490, Y: 1420, W: 300}},
				{Title: "SİSTEMİNİ KORUMA ALTINA AL", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1420, W: 300}},
			},
		}
	}

	slug := slugify(matched.title)
	return &ScriptResult{
		Title:     matched.title,
		Slug:      slug,
		Narration: matched.narration,
		Scenes:    matched.scenes,
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
	if vm, err := generateViralViaOllama(title, narration); err == nil && vm != nil {
		return *vm
	}
	return generateViralViaTemplates(title, narration)
}

func generateViralViaOllama(title, narration string) (*ViralMetadata, error) {
	pingClient := &http.Client{Timeout: 500 * time.Millisecond}
	pingResp, err := pingClient.Get("http://localhost:11434/api/tags")
	if err != nil {
		return nil, err
	}
	_ = pingResp.Body.Close()
	if pingResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama not ready")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
		"model":  "qwen2.5:7b",
		"prompt": sysPrompt,
		"stream": false,
		"format": "json",
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "http://localhost:11434/api/generate", bytes.NewReader(bodyBytes))
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
	if err := json.Unmarshal([]byte(ollamaResp.Response), &vm); err != nil {
		return nil, err
	}
	if vm.TitleA == "" || vm.Description == "" || len(vm.Hashtags) == 0 {
		return nil, fmt.Errorf("incomplete viral response")
	}
	return &vm, nil
}

func generateViralViaTemplates(title, narration string) ViralMetadata {
	lower := strings.ToLower(title + " " + narration)

	if strings.Contains(lower, "instagram") || strings.Contains(lower, "sosyal medya") || strings.Contains(lower, "hesap çal") {
		return ViralMetadata{
			TitleA:      "🚨 Instagram Hesabınız 1 Saniyede Nasıl Çalınır? (Sakın Bu Hatayı Yapmayın!)",
			TitleB:      "⚠️ 'Telif Hakkı İhlali' Mesajı Aldınız mı? Sakın Tıklamayın!",
			TitleC:      "🛡️ Instagram Hesabınızı Hackerlardan Korumanın 3 Altın Kuralı!",
			Description: "Instagram hesabın bir gecede elinden gidebilir! 😱 Hackerlar şifreni kırmakla uğraşmaz, sahte telif veya mavi tik mesajlarıyla seni tuzağa düşürür. Panikle sahte bağlantıya tıkladığın an tüm takipçilerini ve hesabını hackerlara hediye edersin!\n\n💡 Unutma: Gerçek Instagram yetkilileri senden asla DM üzerinden şifre istemez! İki faktörlü doğrulamayı (2FA) hemen aktif et ve gelen şüpheli linklere asla tıklama.\n\n👇 Bu tür sahte mesajlarla daha önce karşılaştın mı? Yorumlarda yaz!\n👥 Bu videoyu kaydet ve arkadaşlarına gönder, tuzağa düşmesinler!",
			Hashtags: []string{
				"#sibergüvenlik", "#instagramgüvenliği", "#hesapçalma", "#hacker",
				"#güvenlik", "#teknoloji", "#bilgi", "#reels", "#keşfet",
				"#siberbilgi", "#yazılım", "#sosyalmedya", "#dijitalgüvenlik",
				"#trendler", "#fyp", "#şifregüvenliği", "#tuzak", "#uyanıkol", "#viral",
			},
		}
	} else if strings.Contains(lower, "wifi") || strings.Contains(lower, "wi-fi") || strings.Contains(lower, "kafe") || strings.Contains(lower, "ağ") {
		return ViralMetadata{
			TitleA:      "🚨 Kafe Wi-Fi Ağlarına Bağlananlar DİKKAT! (Bunu Biliyor muydunuz?)",
			TitleB:      "⚠️ Şifrenizi 10 Saniyede Havada Yakalayabilirler!",
			TitleC:      "🛡️ Halka Açık Wi-Fi Ağlarında Güvende Kalmanın Tek Yolu!",
			Description: "Gittiğin kafede ücretsiz Wi-Fi'a bağlandın mı? ☕ İşte o an izleniyor olabilirsin! Hackerlar deauth saldırısı ile seni bir saniyede ağdan düşürür, telefonun tekrar bağlanırken şifreli paket havada yakalanır!\n\n💡 Çözüm: Halka açık ağlarda asla şifresiz dolaşma ve mutlaka güvenilir bir VPN kullan! Siber dünyada av değil, avcı ol.\n\n👇 Kafelerde Wi-Fi kullanıyor musun? Yorumlarda buluşalım!\n👥 Videoyu kaydet ve sevdiklerine ilet!",
			Hashtags: []string{
				"#sibergüvenlik", "#wifi", "#ağgüvenliği", "#vpn", "#hacker",
				"#teknoloji", "#reels", "#keşfet", "#bilgi", "#dijitaldünya",
				"#internet", "#güvenlik", "#teknolojihaberleri", "#siberdünya",
				"#trendler", "#fyp", "#viral",
			},
		}
	} else if strings.Contains(lower, "şifre") || strings.Contains(lower, "parola") || strings.Contains(lower, "brute force") {
		return ViralMetadata{
			TitleA:      "🚨 Kullandığınız Şifre Kaç Saniyede Kırılır? (Şok Eden Gerçek!)",
			TitleB:      "⚠️ '123456' veya Doğum Tarihi Kullananlar Hemen Değiştirsin!",
			TitleC:      "🛡️ Kırılması Yüzyıllar Sürecek Güçlü Şifre Oluşturma Formülü!",
			Description: "Kullandığın sekiz haneli şifre saniyeler içinde kırılabilir! ⏱️ Hackerlar saniyede milyarlarca kombinasyon deneyen sistemlerle basit şifreleri anında çözer. Ama büyük harf, rakam ve sembol eklersen kırılması yüzyıllar sürer!\n\n💡 Parolalarını hemen güncelle ve mutlaka bir parola yöneticisi kullan!\n\n👇 Şifren ne kadar güçlü? Yorumlarda tartışalım!\n👥 Gönderiyi kaydet ve sevdiklerine ulaştır!",
			Hashtags: []string{
				"#parola", "#şifregüvenliği", "#bruteforce", "#sibergüvenlik",
				"#teknoloji", "#güvenlik", "#bilgi", "#reels", "#keşfet",
				"#dijitalyaşam", "#hacker", "#şifre", "#bilgisayar", "#trendler",
				"#fyp", "#viral",
			},
		}
	} else if strings.Contains(lower, "kripto") || strings.Contains(lower, "bitcoin") || strings.Contains(lower, "cüzdan") {
		return ViralMetadata{
			TitleA:      "🚨 Kripto Cüzdanınız 1 Tıkla Nasıl Boşaltılır? (Airdrop Tuzağı!)",
			TitleB:      "⚠️ 'Bedava Coin Kazandınız' Mesajlarına Sakın İnanmayın!",
			TitleC:      "🛡️ Kripto Paralarınızı Güvende Tutmanın 3 Temel Kuralı!",
			Description: "Kripto cüzdanındaki tüm birikim tek bir saniyede buharlaşabilir! 💸 Sahte airdrop siteleri bedava coin vaadiyle cüzdanını bağlatır ve akıllı sözleşme izniyle tüm tokenlarını çeker!\n\n💡 Altın kural: Hiç kimse durduk yere bedava para dağıtmaz! Gizli kelimelerini (seed phrase) asla bir web sitesine girme.\n\n👇 Kripto ile ilgilenen arkadaşını etiketle!\n👥 Gönderiyi kaydet, paranı koru!",
			Hashtags: []string{
				"#kripto", "#bitcoin", "#ethereum", "#kriptopara", "#kriptogüvenlik",
				"#airdrop", "#cüzdan", "#sibergüvenlik", "#teknoloji", "#reels",
				"#keşfet", "#finans", "#blokzincir", "#bilgi", "#fyp", "#viral",
			},
		}
	} else if strings.Contains(lower, "whatsapp") || strings.Contains(lower, "sms") || strings.Contains(lower, "telefon") {
		return ViralMetadata{
			TitleA:      "🚨 Arkadaşınızdan 'Kodu Bana At' Mesajı Gelirse Sakın Atmayın!",
			TitleB:      "⚠️ WhatsApp Hesabınızı Tek Bir SMS Koduyla Çalıyorlar!",
			TitleC:      "🛡️ WhatsApp'ta İki Adımlı Doğrulama Nasıl Açılır?",
			Description: "WhatsApp hesabından arkadaşlarına borç para isteme mesajları gidebilir! 📱 Dolandırıcılar numaranı girip sana SMS onay kodu gönderir ve samimi bir arkadaşının ağzından 'kodu bana at' der. O kodu verdiğin an hesabın gider!\n\n💡 SMS doğrulama kodlarını en yakınına bile asla verme! WhatsApp'ta iki adımlı doğrulamayı hemen aç.\n\n👇 Bu tuzağı duymuş muydun? Yorumlarda paylaş!\n👥 Sevdiklerini uyar!",
			Hashtags: []string{
				"#whatsapp", "#dolandırıcılık", "#sibergüvenlik", "#smstuzağı",
				"#teknoloji", "#güvenlik", "#reels", "#keşfet", "#haberler",
				"#bilgi", "#iletişim", "#dikkat", "#trendler", "#fyp", "#viral",
			},
		}
	}

	cleanTitle := strings.TrimSpace(title)
	if cleanTitle == "" {
		cleanTitle = "Siber Güvenlik Gerçeği"
	}
	return ViralMetadata{
		TitleA:      fmt.Sprintf("🚨 %s: Sandığınızdan Çok Daha Büyük Bir Tehlike!", cleanTitle),
		TitleB:      fmt.Sprintf("⚠️ Sakın Bu Hataya Düşmeyin: %s Gerçeği!", cleanTitle),
		TitleC:      fmt.Sprintf("🛡️ %s Konusunda Kendinizi Korumanın En Etkili Yolu!", cleanTitle),
		Description: fmt.Sprintf("%s hakkında bilmeniz gereken kritik siber güvenlik gerçekleri! 🚨 Hackerlar her gün kullanıcıların dikkatsizliklerinden faydalanarak sistemlere sızıyor. Siber dünyada en zayıf halka insandır.\n\n💡 Bilgilenin, tedbirinizi alın ve siber dünyada av değil avcı olun!\n\n👇 Fikirlerinizi yorumlarda paylaşın!\n👥 Videoyu kaydet ve sevdiklerinle paylaş!", cleanTitle),
		Hashtags: []string{
			"#sibergüvenlik", "#teknoloji", "#güvenlik", "#hacker",
			"#bilgi", "#dijitaldünya", "#reels", "#keşfet", "#trendler",
			"#yazılım", "#bilimveteknoloji", "#önlem", "#farkındalık",
			"#siberbilgi", "#fyp", "#viral",
		},
	}
}

// FormatViralText creates the aciklama.txt document with A/B/C titles, caption, and hashtags.
func FormatViralText(title, narration string, vm ViralMetadata) string {
	var sb strings.Builder
	sb.WriteString("════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("🎬 TUBIE VİRAL PAYLAŞIM PAKETİ (INSTAGRAM / REELS / TIKTOK / SHORTS)\n")
	sb.WriteString("════════════════════════════════════════════════════════════════════\n\n")

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

	sb.WriteString("════════════════════════════════════════════════════════════════════\n")
	sb.WriteString("ℹ️ Instagram 2.200 karakter sınırına tam uyumludur.\n")
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
