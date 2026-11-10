package tubie

// Preset represents a cyber security viral video script and configuration.
type Preset struct {
	ID        string            `json:"id"`
	Title     string            `json:"title"`
	Slug      string            `json:"slug"`
	Narration string            `json:"narration"`
	Scenes    []SceneDefinition `json:"scenes"`
}

// CyberPresets contains the 4 pre-configured cyber security viral video topics.
var CyberPresets = map[string]Preset{
	"1": {
		ID:        "1",
		Title:     "Wi-Fi Şifreniz 10 Saniyede Nasıl Kırılır?",
		Slug:      "wifi_sifresi_nasil_kirilir",
		Narration: "Şu an bağlandığın kafe Wi-Fi ağı seni izliyor olabilir! Hackerlar şifreni tahmin etmeye çalışmaz. Deauth saldırısıyla seni bir saniyede ağdan düşürürler. Telefonun tekrar bağlanırken şifreli paket havada yakalanır! Buna dört yollu el sıkışma denir. Hacker bu paketi güçlü bir bilgisayarda saniyeler içinde çözer. Çözüm çok basit: Halka açık ağlarda asla şifresiz gezme ve mutlaka VPN kullan! Siber dünyada av değil, avcı ol. Takipte kal!",
		Scenes: []SceneDefinition{
			{Title: "KAFE WI-FI AĞI GÜVENLİ Mİ?", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-wifi.png", X: 490, Y: 1420, W: 300}},
			{Title: "BİR SANİYEDE AĞDAN DÜŞTÜN!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 490, Y: 1420, W: 300}, Arrow: true},
			{Title: "ŞİFRELİ PAKET HAVADA YAKALANDI", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 490, Y: 1420, W: 300}},
			{Title: "DÖRT YOLLU EL SIKIŞMA ÇÖZÜLDÜ", Color: "#FF3B30", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 490, Y: 1420, W: 300}},
			{Title: "HER ZAMAN VPN KULLAN!", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1420, W: 300}},
			{Title: "AV DEĞİL, AVCI OL!", Color: "#00F0FF", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-padlock-unlocked.png", X: 490, Y: 1420, W: 300}},
		},
	},
	"2": {
		ID:        "2",
		Title:     "Şifreniz Ne Kadar Sürede Kırılır? (Brute Force Gerçeği)",
		Slug:      "sifren_nasil_kirilir_brute_force",
		Narration: "Kullandığın sekiz haneli şifre saniyeler içinde kırılabilir! Çoğu insan hala doğum tarihi veya bir iki üç dört beş altı kullanıyor. Hackerlar saniyede milyarlarca kombinasyon deneyen sistemler kullanır. Buna kaba kuvvet saldırısı denir. Sadece küçük harfli bir şifre anında çözülür. Ama içine büyük harf, rakam ve özel karakter eklersen? Kırılması yüzyıllar sürer! Şifreni hemen güncelle, parolalarını bir yöneticiye emanet et. Siber güvenlik şakaya gelmez!",
		Scenes: []SceneDefinition{
			{Title: "ŞİFRENİZ NE KADAR GÜVENLİ?", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 490, Y: 1420, W: 300}},
			{Title: "SANİYEDE MİLYARLARCA DENEME", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 490, Y: 1420, W: 300}, Arrow: true},
			{Title: "BASİT ŞİFRELER ANINDA ÇÖZÜLÜR!", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 490, Y: 1420, W: 300}},
			{Title: "GÜÇLÜ PAROLA FORMÜLÜ", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1420, W: 300}},
			{Title: "PAROLA YÖNETİCİSİ KULLAN", Color: "#00F0FF", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-padlock-unlocked.png", X: 490, Y: 1420, W: 300}},
			{Title: "SİBER GÜVENLİK ŞAKAYA GELMEZ!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1420, W: 300}},
		},
	},
	"3": {
		ID:        "3",
		Title:     "SQL Injection Nedir ve Nasıl Çalışır?",
		Slug:      "sql_injection_nasil_calisir",
		Narration: "Tek bir tırnak işaretiyle koca bir şirketin veritabanı boşaltılabilir mi? Evet! Buna SQL Injection denir. Web sitesine giriş yaparken kullanıcı adına 'veya bir eşittir bir' yazarsan ne olur? Sistem bunu kod zanneder ve şifre kontrolünü tamamen atlar! Hacker tek tıkla admin yetkisini ele geçirir, tüm kullanıcı verilerini çalar. Yazılımcılar dikkat: Kullanıcıdan gelen hiçbir veriye güvenmeyin, parametrik sorgular kullanın. Kodunu koru, geleceğini koru!",
		Scenes: []SceneDefinition{
			{Title: "TEK BİR TIRNAK İŞARETİYLE VERİTABANI!", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 490, Y: 1420, W: 300}},
			{Title: "VEYA BİR EŞİTTİR BİR HİLESİ", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 490, Y: 1420, W: 300}, Arrow: true},
			{Title: "GİRİŞ KONTROLÜ ATLANDI!", Color: "#FFCC00", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 490, Y: 1420, W: 300}},
			{Title: "ADMİN YETKİSİ ELE GEÇİRİLDİ", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-padlock-unlocked.png", X: 490, Y: 1420, W: 300}},
			{Title: "PARAMETRİK SORGU KULLANIN", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1420, W: 300}},
			{Title: "KODUNU KORU, GÜVENLE KAL!", Color: "#00F0FF", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1420, W: 300}},
		},
	},
	"4": {
		ID:        "4",
		Title:     "Oltalama (Phishing) Tuzağına Nasıl Düşürülürsünüz?",
		Slug:      "phishing_oltalama_tuzagi",
		Narration: "Hesabınız askıya alındı, hemen tıklayın mesajı aldınız mı? Sakın tıklamayın! Bu bir oltalama tuzağıdır. Hackerlar bankanın veya popüler sitelerin birebir kopyasını yapar. Panikle linke tıklar ve şifrenizi girersiniz. Tebrikler, hesabınızı kendi ellerinizle teslim ettiniz! Gerçek kurumlar sizden asla mesajla şifre istemez. Şüpheli linklere tıklama, gönderici adresini mutlaka kontrol et. Tuzağa düşme, uyanık ol!",
		Scenes: []SceneDefinition{
			{Title: "HESABINIZ ASKIYA ALINDI MESAJI!", Color: "#FF3B30", SFX: "alert", Prop: &SceneProp{Path: "@b1/cyber-laptop.png", X: 490, Y: 1420, W: 300}},
			{Title: "SAHTE GİRİŞ SAYFASI KOPYASI", Color: "#FFCC00", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-terminal.png", X: 490, Y: 1420, W: 300}, Arrow: true},
			{Title: "ŞİFRENİ KENDİ ELLERİNLE VERDİN!", Color: "#FF3B30", SFX: "glitch", Prop: &SceneProp{Path: "@b1/cyber-skull.png", X: 490, Y: 1420, W: 300}},
			{Title: "GERÇEK KURUMLAR ŞİFRE İSTEMEZ", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1420, W: 300}},
			{Title: "GÖNDERİCİ ADRESİNİ KONTROL ET", Color: "#00F0FF", SFX: "keyboard", Prop: &SceneProp{Path: "@b1/cyber-wifi.png", X: 490, Y: 1420, W: 300}},
			{Title: "TUZAĞA DÜŞME, UYANIK OL!", Color: "#00FF66", SFX: "lock", Prop: &SceneProp{Path: "@b1/cyber-shield.png", X: 490, Y: 1420, W: 300}},
		},
	},
}

// PresetList returns the presets in ordered slice.
func PresetList() []Preset {
	keys := []string{"1", "2", "3", "4"}
	list := make([]Preset, 0, len(keys))
	for _, k := range keys {
		if p, ok := CyberPresets[k]; ok {
			list = append(list, p)
		}
	}
	return list
}
