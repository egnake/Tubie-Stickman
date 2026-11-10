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
Bu konuda 30 saniyelik çok akıcı, heyecan verici, yüksek tempolu ve bol aksiyonlu Türkçe seslendirme metni yaz.
Hackerların nasıl tuzak kurduğunu ve kullanıcının nasıl korunacağını anlat.
Tam olarak 6 farklı sahneden oluşan dinamik bir görsel akış planla.
Yanıtını kesinlikle SADECE geçerli bir JSON olarak ver:
{
  "title": "Kısa ve çarpıcı video başlığı",
  "narration": "60-80 kelimelik yüksek tempolu Türkçe seslendirme metni",
  "scenes": [
    {"title": "Sahne 1 (Kanca / Şok)", "sfx": "keyboard", "color": "#00F0FF"},
    {"title": "Sahne 2 (Tuzak Mesajı)", "sfx": "alert", "color": "#FF3B30"},
    {"title": "Sahne 3 (Sızma / Çalınma)", "sfx": "glitch", "color": "#FFCC00"},
    {"title": "Sahne 4 (Hacker Terminali)", "sfx": "keyboard", "color": "#E056FD"},
    {"title": "Sahne 5 (2FA Kalkanı / Çözüm)", "sfx": "lock", "color": "#00FF66"},
    {"title": "Sahne 6 (Zafer / Takip Çağrısı)", "sfx": "chime", "color": "#00F0FF"}
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
			title:     "İnstagram Hesabınız Nasıl Çalınır?",
			narration: "Instagram hesabın bir gecede elinden gidebilir! Hackerlar şifreni kırmakla uğraşmaz. 'Telif hakkı ihlali' veya 'Mavi tik kazandınız' diye sahte bir DM atarlar. Panikle linke tıklar ve sahte giriş sayfasına şifreni yazarsın. Tebrikler, hesabını ve tüm takipçilerini hacker'a hediye ettin! Gerçek Instagram asla DM'den şifre istemez. İki faktörlü doğrulamayı hemen aç ve şüpheli linklere tıklama! Uyanık kal, hesabını koru!",
			scenes: []SceneDefinition{
				{Title: "SAHTE DM TUZAĞI!", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 490, Y: 1380, W: 300}},
				{Title: "ŞİFREN ELE GEÇİRİLDİ!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 490, Y: 1380, W: 300}, Arrow: true},
				{Title: "HESAP KONTROLÜ KAYBEDİLDİ", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 490, Y: 1380, W: 300}},
				{Title: "GERÇEK INSTAGRAM ŞİFRE İSTEMEZ", Color: "#E056FD", SFX: "paper", Prop: &SceneProp{Path: "@b1/idea-bulb-crayon.png", X: 490, Y: 1380, W: 300}},
				{Title: "2FA İLE HESABINI KORU!", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1380, W: 300}},
				{Title: "UYANIK KAL, GÜVENDE OL!", Color: "#00F0FF", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 490, Y: 1380, W: 300}},
			},
		}
	} else if strings.Contains(lower, "kripto") || strings.Contains(lower, "bitcoin") || strings.Contains(lower, "cüzdan") {
		matched = &tpl{
			title:     "Kripto Cüzdanınız Nasıl Sıfırlanır?",
			narration: "Kripto cüzdanındaki tüm birikim tek bir saniyede buharlaşabilir! Sahte airdrop siteleri ve dolandırıcı Telegram grupları sana bedava coin vaat eder. Cüzdanını bağladığın an, akıllı sözleşme izniyle tüm tokenlarını çekerler! Altın kural: Hiç kimse durduk yere bedava para dağıtmaz. Gizli kelimelerini asla bir web sitesine yazma ve her sözleşme onayını iki kez kontrol et! Parana sahip çık!",
			scenes: []SceneDefinition{
				{Title: "BEDAVA COIN TUZAĞI!", Color: "#FFCC00", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 490, Y: 1380, W: 300}},
				{Title: "AKILLI SÖZLEŞME ONAYI", Color: "#FF3B30", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 490, Y: 1380, W: 300}, Arrow: true},
				{Title: "CÜZDAN BOŞALTILDI!", Color: "#FF3B30", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 490, Y: 1380, W: 300}},
				{Title: "BEDAVA PARA DİYE BİR ŞEY YOK!", Color: "#00F0FF", SFX: "paper", Prop: &SceneProp{Path: "@b1/idea-bulb-crayon.png", X: 490, Y: 1380, W: 300}},
				{Title: "KELİMELERİNİ ASLA PAYLAŞMA", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 490, Y: 1380, W: 300}},
				{Title: "PARANA VE CÜZDANINA SAHİP ÇIK!", Color: "#00FF66", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1380, W: 300}},
			},
		}
	} else if strings.Contains(lower, "whatsapp") || strings.Contains(lower, "telefon") || strings.Contains(lower, "sms") {
		matched = &tpl{
			title:     "WhatsApp Hesabınız Nasıl Ele Geçirilir?",
			narration: "WhatsApp hesabından arkadaşlarına borç para isteme mesajları gidebilir! Hacker numaranı sisteme yazar ve sana bir SMS onay kodu gelir. Hemen ardından samimi bir arkadaşından mesaj alırsın: 'Kodu yanlışlıkla sana yolladım, bana atar mısın?' O kodu verdiğin anda hesabın gider! SMS doğrulama kodlarını en yakınına bile asla verme. WhatsApp'ta iki adımlı doğrulamayı hemen etkinleştir!",
			scenes: []SceneDefinition{
				{Title: "BEKLENMEDİK ONAY KODU!", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-wifi.png", X: 490, Y: 1380, W: 300}},
				{Title: "KODU VERDİĞİN AN BİTTİN!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 490, Y: 1380, W: 300}, Arrow: true},
				{Title: "HESAP BAŞKASINA GEÇTİ", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 490, Y: 1380, W: 300}},
				{Title: "ARKADAŞINDAN GELEN SAHTE MESAJ", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 490, Y: 1380, W: 300}},
				{Title: "İKİ ADIMLI DOĞRULAMAYI AÇ", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1380, W: 300}},
				{Title: "DOĞRULAMA KODUNU ASLA VERME!", Color: "#00FF66", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 490, Y: 1380, W: 300}},
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
			title:     title,
			narration: fmt.Sprintf("Dikkat! %s konusu sandığından çok daha kritik bir siber güvenlik riski taşıyor olabilir. Hackerlar her gün binlerce kullanıcının dikkatsizliğinden faydalanarak sistemlere sızıyor. Basit bir dikkatsizlik, tüm kişisel verilerinin ve parolalarının açığa çıkmasına neden olabilir. Unutma: Siber güvenlikte en zayıf halka insandır. Güçlü parolalar kullan, çift faktörlü korumayı aç ve daima tetikte kal! Takipte ol, güvenle kal!", cleanSubject),
			scenes: []SceneDefinition{
				{Title: "GÖRÜNMEYEN TEHLİKE!", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 490, Y: 1380, W: 300}},
				{Title: "SİSTEME SIZMA ANI!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 490, Y: 1380, W: 300}, Arrow: true},
				{Title: "VERİLERİN AÇIĞA ÇIKMASI", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 490, Y: 1380, W: 300}},
				{Title: "EN ZAYIF HALKA: DİKKATSİZLİK", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-bug.png", X: 490, Y: 1380, W: 300}},
				{Title: "SİSTEMİNİ KORUMA ALTINA AL", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1380, W: 300}},
				{Title: "AV DEĞİL, AVCI OL! TAKİP ET", Color: "#00FF66", SFX: "chime", Prop: &SceneProp{Path: "@b1/idea-bulb-crayon.png", X: 490, Y: 1380, W: 300}},
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

// FixedYouTubeTags contains the mandatory, high-converting YouTube tags for cybersecurity videos.
const FixedYouTubeTags = "hacker,hack,hackers,hacks,atıl samancıoğlu,can değer,mdisec,para kazanma,etik hacker olma kursu,etik hacker,panel,cc,active directory,redteam,blueteam,pentest,penetration test,analist,soc analisti,soc,ioc,iot,bugbounty,bug bount,etik hacker olma,hacker olma kursu,siber tehdit,siber güvenlik,cyber security,cyber sec,bilgi güvenliği,mit,btk,yazılım,bilişim,bilgisayar,network,veri güvenliği,bilişim güvenliği,siyah şapkalı hacker,hackerlık,darkweb,deepweb,osint"

func generateViralViaTemplates(title, narration string) ViralMetadata {
	lower := strings.ToLower(title + " " + narration)

	if strings.Contains(lower, "instagram") || strings.Contains(lower, "sosyal medya") || strings.Contains(lower, "hesap çal") {
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
		TitleA:      fmt.Sprintf("🚨 %s: Sandığınızdan Çok Daha Büyük Bir Tehlike!", cleanTitle),
		TitleB:      fmt.Sprintf("⚠️ Sakın Bu Hataya Düşmeyin: %s Gerçeği!", cleanTitle),
		TitleC:      fmt.Sprintf("🛡️ %s Konusunda Kendinizi Korumanın En Etkili Yolu!", cleanTitle),
		Description: fmt.Sprintf("%s hakkında bilmeniz gereken kritik siber güvenlik gerçekleri! 🚨\n\nHackerlar her gün kullanıcıların dikkatsizliklerinden faydalanarak sistemlere sızıyor. Siber dünyada en zayıf halka insandır.\n\n⚠️ DİJİTAL DÜNYADAKİ EN BÜYÜK RİSKLER:\n1️⃣ Oltalama (Phishing) Bildirimleri ve sahte e-postalar.\n2️⃣ Crackli oyun ve korsan yazılımlara gizlenmiş truva atları.\n3️⃣ Darkweb'e sızdırılan parola ve kimlik verileri.\n\n🛡️ DİJİTAL KALKANINIZI GÜÇLENDİRİN:\n✅ 1. Bilmediğiniz hiçbir dosya veya e-posta ekini indirmeyin.\n✅ 2. Bütün kritik hesaplarınızda 2FA koruması açık olsun.\n✅ 3. Lisanssız korsan yazılımlardan uzak durun.\n✅ 4. Dijital dünyada av değil, bilinçli bir avcı olun!\n\n👇 Siz dijital güvenliğinizi nasıl sağlıyorsunuz? Yorumlarda tartışalım!\n👥 Gönderiyi kaydedin ve sevdiklerinizle paylaşın!", cleanTitle),
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
