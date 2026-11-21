package tubie

// Preset represents a cyber security viral video script and configuration.
type Preset struct {
	ID          string            `json:"id"`
	Category    string            `json:"category,omitempty"`
	Title       string            `json:"title"`
	Summary     string            `json:"summary,omitempty"`
	HookPreview string            `json:"hook_preview,omitempty"`
	Slug        string            `json:"slug"`
	Narration   string            `json:"narration"`
	Scenes      []SceneDefinition `json:"scenes"`
}

// CyberPresets contains 10 pre-configured cyber security viral video topics.
var CyberPresets = map[string]Preset{
	"1": {
		ID:          "1",
		Category:    "Ağ, Wi-Fi & Kablosuz",
		Title:       "Wi-Fi Şifreniz 10 Saniyede Nasıl Kırılır?",
		Summary:     "Deauth ve Evil Twin kafe Wi-Fi saldırıları",
		HookPreview: "Şu an bağlandığın kafe Wi-Fi ağı seni izliyor olabilir!",
		Slug:        "wifi_sifresi_nasil_kirilir",
		Narration:   "Şu an bağlandığın kafe Wi-Fi ağı seni izliyor olabilir! Hackerlar şifreni tahmin etmeye çalışmaz. Deauth saldırısıyla seni bir saniyede ağdan düşürürler. Telefonun tekrar bağlanırken şifreli paket havada yakalanır! Buna dört yollu el sıkışma denir. Hacker bu paketi güçlü bir bilgisayarda saniyeler içinde çözer. Çözüm çok basit: Halka açık ağlarda asla şifresiz gezme ve mutlaka VPN kullan! Siber dünyada av değil, avcı ol. Takipte kal!",
		Scenes: []SceneDefinition{
			{Title: "KAFE WI-FI AĞI GÜVENLİ Mİ?", VisualType: "notification", Mood: "hook", Camera: "punch_in", Callout: "DİKKAT!", Color: "#38BDF8", SFX: "notification", Prop: &SceneProp{Path: "@b1/cyber-wifi.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"sender": "Ücretsiz Kafe Wi-Fi", "body": "Otomatik bağlanıldı: Free_Cafe_WiFi (Şifresiz)", "badge": "AĞ BAĞLANDI"}},
			{Title: "BİR SANİYEDE AĞDAN DÜŞTÜN!", VisualType: "network_mitm", Mood: "danger", Camera: "shake", Callout: "DEAUTH SALDIRISI!", Color: "#F43F5E", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}, Arrow: true, VisualData: map[string]any{"left": "TELEFON", "middle": "SAHTE AP", "right": "MODEM", "packets": []string{"DEAUTH", "HANDSHAKE", "HASH"}}},
			{Title: "ŞİFRELİ PAKET HAVADA YAKALANDI", VisualType: "terminal", Mood: "breach", Camera: "glitch_cut", Callout: "PAKET YAKALANDI!", Color: "#F59E0B", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"header": "aircrack@kali:~/capture", "lines": []string{"$ aireplay-ng --deauth 10 -a AP", "[+] WPA Handshake captured!", "[*] Cracking PMKID with GPU...", "[!] KEY FOUND IN 8.4 SECONDS!"}}},
			{Title: "DÖRT YOLLU EL SIKIŞMA ÇÖZÜLDÜ", VisualType: "threat_meter", Mood: "twist", Camera: "zoom_slow", Callout: "RİSK %96!", Color: "#F43F5E", SFX: "radar", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"level": 96, "label": "TRAFİK AÇIKTA", "safe": false}},
			{Title: "HER ZAMAN VPN KULLAN!", VisualType: "checklist", Mood: "defense", Camera: "punch_in", Callout: "3 NET KURAL!", Color: "#10B981", SFX: "shield", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"items": []string{"Ortak Wi-Fi'de bankaya girme", "Otomatik ağ bağlanmayı kapat", "Her zaman şifreli VPN aç"}}},
			{Title: "AV DEĞİL, AVCI OL!", VisualType: "versus", Mood: "cta", Camera: "wide", Callout: "GÜVENDE KAL!", Color: "#38BDF8", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"leftTitle": "ŞİFRESİZ ORTAK AĞ", "leftText": "Deauth & paket dinleme riski", "rightTitle": "VPN + MOBİL VERİ", "rightText": "Uçtan uca şifreli trafik"}},
		},
	},
	"2": {
		ID:          "2",
		Category:    "Parola & Kimlik",
		Title:       "Şifreniz Ne Kadar Sürede Kırılır? (Brute Force Gerçeği)",
		Summary:     "GPU kaba kuvvet hızı ve 14+ karakter formülü",
		HookPreview: "Kullandığın sekiz haneli şifre saniyeler içinde kırılabilir!",
		Slug:        "sifren_nasil_kirilir_brute_force",
		Narration:   "Kullandığın sekiz haneli şifre saniyeler içinde kırılabilir! Çoğu insan hala doğum tarihi veya bir iki üç dört beş altı kullanıyor. Hackerlar saniyede milyarlarca kombinasyon deneyen sistemler kullanır. Buna kaba kuvvet saldırısı denir. Sadece küçük harfli bir şifre anında çözülür. Ama içine büyük harf, rakam ve özel karakter eklersen? Kırılması yüzyıllar sürer! Şifreni hemen güncelle, parolalarını bir yöneticiye emanet et. Siber güvenlik şakaya gelmez!",
		Scenes: []SceneDefinition{
			{Title: "ŞİFRENİZ NE KADAR GÜVENLİ?", VisualType: "password_crack", Mood: "hook", Camera: "punch_in", Callout: "8 HANE YETMEZ!", Color: "#38BDF8", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 390, Y: 1040, W: 500}},
			{Title: "SANİYEDE MİLYARLARCA DENEME", VisualType: "terminal", Mood: "danger", Camera: "shake", Callout: "BRUTE FORCE!", Color: "#F43F5E", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}, Arrow: true, VisualData: map[string]any{"header": "hashcat@gpu-rig:~", "lines": []string{"$ hashcat -m 0 -a 3 rockyou.txt", "[*] Speed: 94.2 GH/s (RTX Cluster)", "[+] Match: ahmet1998 (0.4 sec)", "[!] PASSWORD CRACKED!"}}},
			{Title: "BASİT ŞİFRELER ANINDA ÇÖZÜLÜR!", VisualType: "threat_meter", Mood: "breach", Camera: "glitch_cut", Callout: "0.4 SANİYE!", Color: "#F59E0B", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"level": 99, "label": "ANINDA KIRILDI", "safe": false}},
			{Title: "GÜÇLÜ PAROLA FORMÜLÜ", VisualType: "versus", Mood: "twist", Camera: "zoom_slow", Callout: "YÜZYILLAR SÜRER!", Color: "#10B981", SFX: "radar", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"leftTitle": "8 KÜÇÜK HARF", "leftText": "0.4 saniyede çözülür", "rightTitle": "14+ KARAKTER", "rightText": "Kırılması yüzyıllar sürer"}},
			{Title: "PAROLA YÖNETİCİSİ KULLAN", VisualType: "checklist", Mood: "defense", Camera: "punch_in", Callout: "3 ALTIN KURAL!", Color: "#38BDF8", SFX: "shield", Prop: &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"items": []string{"En az 14 karakter parola kur", "Her siteye farklı şifre ver", "Parola yöneticisi + Passkey aç"}}},
			{Title: "SİBER GÜVENLİK ŞAKAYA GELMEZ!", VisualType: "threat_meter", Mood: "cta", Camera: "wide", Callout: "%100 KORUMA!", Color: "#10B981", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"level": 100, "label": "PAROLA ZIRHLI", "safe": true}},
		},
	},
	"3": {
		ID:          "3",
		Category:    "Web, Veri & Kod Güvenliği",
		Title:       "SQL Injection Nedir ve Nasıl Çalışır?",
		Summary:     "Tek tırnak işaretiyle yetki aşımı ve parametrik sorgu savunması",
		HookPreview: "Tek bir tırnak işaretiyle koca bir şirketin veritabanı boşaltılabilir mi?",
		Slug:        "sql_injection_nasil_calisir",
		Narration:   "Tek bir tırnak işaretiyle koca bir şirketin veritabanı boşaltılabilir mi? Evet! Buna SQL Injection denir. Web sitesine giriş yaparken kullanıcı adına 'veya bir eşittir bir' yazarsan ne olur? Sistem bunu kod zanneder ve şifre kontrolünü tamamen atlar! Hacker tek tıkla admin yetkisini ele geçirir, tüm kullanıcı verilerini çalar. Yazılımcılar dikkat: Kullanıcıdan gelen hiçbir veriye güvenmeyin, parametrik sorgular kullanın. Kodunu koru, geleceğini koru!",
		Scenes: []SceneDefinition{
			{Title: "TEK BİR TIRNAK İŞARETİYLE VERİTABANI!", VisualType: "url_inspect", Mood: "hook", Camera: "punch_in", Callout: "GİRİŞ KUTUSU!", Color: "#38BDF8", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"fakeUrl": "user: admin' OR 1=1 --", "realUrl": "Parametrik Sorgu", "stamp": "SQL INJECTION!"}},
			{Title: "VEYA BİR EŞİTTİR BİR HİLESİ", VisualType: "terminal", Mood: "danger", Camera: "shake", Callout: "' OR 1=1 --", Color: "#F43F5E", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 390, Y: 1040, W: 500}, Arrow: true, VisualData: map[string]any{"header": "sqlmap@kali:~", "lines": []string{"SELECT * FROM users WHERE", "user='' OR 1=1 -- ' AND pass=''", "[+] QUERY EVALUATED AS TRUE!", "[!] LOGGED IN AS ADMIN!"}}},
			{Title: "GİRİŞ KONTROLÜ ATLANDI!", VisualType: "network_mitm", Mood: "breach", Camera: "glitch_cut", Callout: "VERİ ÇEKİLİYOR!", Color: "#F59E0B", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"left": "GİRİŞ FORMU", "middle": "SQL SORGUSU", "right": "VERİTABANI", "packets": []string{"' OR 1=1", "USERS", "HASHES"}}},
			{Title: "ADMİN YETKİSİ ELE GEÇİRİLDİ", VisualType: "threat_meter", Mood: "twist", Camera: "zoom_slow", Callout: "TAM İHLAL!", Color: "#F43F5E", SFX: "radar", Prop: &SceneProp{Path: "@b1/cyber-padlock-unlocked.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"level": 98, "label": "VERİTABANI AÇIK", "safe": false}},
			{Title: "PARAMETRİK SORGU KULLANIN", VisualType: "checklist", Mood: "defense", Camera: "punch_in", Callout: "GÜVENLİ KOD!", Color: "#10B981", SFX: "shield", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"items": []string{"Prepared Statements kullan", "En düşük DB yetkisini ver", "Ham SQL hatalarını gizle"}}},
			{Title: "KODUNU KORU, GÜVENLE KAL!", VisualType: "versus", Mood: "cta", Camera: "wide", Callout: "ZERO SQLi!", Color: "#38BDF8", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"leftTitle": "STRING BİRLEŞTİRME", "leftText": "Saldırgan sorguyu değiştirir", "rightTitle": "PARAMETRİK BINDING", "rightText": "Girdi yalnızca veri kalır"}},
		},
	},
	"4": {
		ID:          "4",
		Category:    "Oltalama & Sosyal Mühendislik",
		Title:       "Oltalama (Phishing) Tuzağına Nasıl Düşürülürsünüz?",
		Summary:     "Sahte giriş sayfası, aciliyet baskısı ve gönderici doğrulama",
		HookPreview: "Hesabınız askıya alındı, hemen tıklayın mesajı aldınız mı?",
		Slug:        "phishing_oltalama_tuzagi",
		Narration:   "Hesabınız askıya alındı, hemen tıklayın mesajı aldınız mı? Sakın tıklamayın! Bu bir oltalama tuzağıdır. Hackerlar bankanın veya popüler sitelerin birebir kopyasını yapar. Panikle linke tıklar ve şifrenizi girersiniz. Tebrikler, hesabınızı kendi ellerinizle teslim ettiniz! Gerçek kurumlar sizden asla mesajla şifre istemez. Şüpheli linklere tıklama, gönderici adresini mutlaka kontrol et. Tuzağa düşme, uyanık ol!",
		Scenes: []SceneDefinition{
			{Title: "HESABINIZ ASKIYA ALINDI MESAJI!", VisualType: "notification", Mood: "hook", Camera: "punch_in", Callout: "SAHTE UYARI!", Color: "#F43F5E", SFX: "notification", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"sender": "Hesap Güvenliği (?)", "body": "Hesabınız askıya alındı! Hemen doğrulamak için tıklayın.", "badge": "ACİL İŞLEM"}},
			{Title: "SAHTE GİRİŞ SAYFASI KOPYASI", VisualType: "url_inspect", Mood: "danger", Camera: "shake", Callout: "ALAN ADINA BAK!", Color: "#F59E0B", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 390, Y: 1040, W: 500}, Arrow: true, VisualData: map[string]any{"fakeUrl": "https://banka-onay1a.com", "realUrl": "banka.com.tr", "stamp": "SAHTE KLON!"}},
			{Title: "ŞİFRENİ KENDİ ELLERİNLE VERDİN!", VisualType: "terminal", Mood: "breach", Camera: "glitch_cut", Callout: "HESAP GİTTİ!", Color: "#F43F5E", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"header": "phish-panel@attacker:~", "lines": []string{"[+] Victim clicked fake link", "[+] Captured email + password", "[!] Bypassing session token...", ">>> ACCOUNT TAKEN OVER!"}}},
			{Title: "GERÇEK KURUMLAR ŞİFRE İSTEMEZ", VisualType: "threat_meter", Mood: "twist", Camera: "zoom_slow", Callout: "ASLA PANİKLEME!", Color: "#10B981", SFX: "radar", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"level": 92, "label": "MANİPÜLASYON RİSKİ", "safe": false}},
			{Title: "GÖNDERİCİ ADRESİNİ KONTROL ET", VisualType: "checklist", Mood: "defense", Camera: "punch_in", Callout: "DUR. DOĞRULA!", Color: "#38BDF8", SFX: "shield", Prop: &SceneProp{Path: "@b1/cyber-wifi.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"items": []string{"Gönderen adresini harf harf oku", "Link yerine resmi uygulamaya gir", "Authenticator tabanlı 2FA aç"}}},
			{Title: "TUZAĞA DÜŞME, UYANIK OL!", VisualType: "versus", Mood: "cta", Camera: "wide", Callout: "UYANIK OL!", Color: "#10B981", SFX: "chime", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 390, Y: 1040, W: 500}, VisualData: map[string]any{"leftTitle": "LİNKE PANİKLE TIKLA", "leftText": "Sahte sayfaya şifre teslimi", "rightTitle": "DOĞRUDAN UYGULAMA", "rightText": "Sıfır oltalama riski"}},
		},
	},
	"5": {
		ID:          "5",
		Category:    "Yapay Zeka & Deepfake",
		Title:       "3 Saniyelik Sesinle Nasıl Dolandırılırsın?",
		Summary:     "Yapay zeka ses klonlama ve aile parolası savunması",
		HookPreview: "Gece yarısı tanıdık bir sesle gelen acil para talebi gerçek olmayabilir!",
		Slug:        "deepfake_ses_klonlama_tuzagi",
		Narration:   lookupTopicBlueprint("deepfake").narration[StyleHybrid],
		Scenes:      lookupTopicBlueprint("deepfake").scenes,
	},
	"6": {
		ID:          "6",
		Category:    "Kripto & Web3 Güvenliği",
		Title:       "Kripto Cüzdanın Tek Onayla Nasıl Boşaltılır?",
		Summary:     "Sahte airdrop ve sınırsız Approve drainer imzası",
		HookPreview: "Gizli kelimelerini çalmadan tek bir imzayla tüm cüzdan boşaltılabilir!",
		Slug:        "kripto_cuzdan_airdrop_tuzagi",
		Narration:   lookupTopicBlueprint("crypto").narration[StyleHybrid],
		Scenes:      lookupTopicBlueprint("crypto").scenes,
	},
	"7": {
		ID:          "7",
		Category:    "Zararlı Yazılım & Cihaz",
		Title:       "Ücretsiz Crack İndirirken Bilgisayarında Ne Olur?",
		Summary:     "Infostealer bilgi hırsızları ve çerez sızıntısı",
		HookPreview: "Antivirüsü kapat diyen o kurulum 3 saniyede tüm şifrelerini çalar!",
		Slug:        "ucretsiz_crack_infostealer_tehlikesi",
		Narration:   lookupTopicBlueprint("malware").narration[StyleHybrid],
		Scenes:      lookupTopicBlueprint("malware").scenes,
	},
	"8": {
		ID:          "8",
		Category:    "Mobil & Kripto",
		Title:       "WhatsApp SMS Kodu & SIM Swap Hilesi",
		Summary:     "6 haneli SMS kodu hırsızlığı ve iki adımlı PIN kilidi",
		HookPreview: "En yakın arkadaşından gelen 'kodu yanlışlıkla sana attım' mesajına dikkat!",
		Slug:        "whatsapp_sms_kodu_calma_hilesi",
		Narration:   lookupTopicBlueprint("whatsapp").narration[StyleHybrid],
		Scenes:      lookupTopicBlueprint("whatsapp").scenes,
	},
	"9": {
		ID:          "9",
		Category:    "Yapay Zeka & Deepfake",
		Title:       "Yapay Zeka Seni İnternette Ne Kadar Tanıyor?",
		Summary:     "Sohbet geçmişi, model eğitimi kapatma ve kişisel veri hijyeni",
		HookPreview: "Yapay zeka sohbetlerin senin hakkında detaylı bir profil çıkarıyor olabilir!",
		Slug:        "yapay_zeka_dijital_gizlilik",
		Narration:   lookupTopicBlueprint("ai-privacy").narration[StyleHybrid],
		Scenes:      lookupTopicBlueprint("ai-privacy").scenes,
	},
	"10": {
		ID:          "10",
		Category:    "Oltalama & Sosyal Mühendislik",
		Title:       "Sokaktaki QR Kodu Okutmadan Önce Dur!",
		Summary:     "Quishing sahte karekod etiketi ve açık USB şarj riskleri",
		HookPreview: "Otoparktaki karekodun üzerine geceden sahte bir etiket yapıştırılmış olabilir!",
		Slug:        "qr_kod_quishing_tuzagi",
		Narration:   lookupTopicBlueprint("qr").narration[StyleHybrid],
		Scenes:      lookupTopicBlueprint("qr").scenes,
	},
}

// PresetList returns the presets in ordered slice.
func PresetList() []Preset {
	keys := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}
	list := make([]Preset, 0, len(keys))
	for _, k := range keys {
		if p, ok := CyberPresets[k]; ok {
			list = append(list, p)
		}
	}
	return list
}
