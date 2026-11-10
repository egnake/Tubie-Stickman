package tubie

// Preset represents a cyber security viral video script and configuration.
type Preset struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Slug      string `json:"slug"`
	Narration string `json:"narration"`
	Scenes    int    `json:"scenes"`
}

// CyberPresets contains the 4 pre-configured cyber security viral video topics.
var CyberPresets = map[string]Preset{
	"1": {
		ID:        "1",
		Title:     "Wi-Fi Şifreniz 10 Saniyede Nasıl Kırılır?",
		Slug:      "wifi_sifresi_nasil_kirilir",
		Narration: "Şu an bağlandığın kafe Wi-Fi ağı seni izliyor olabilir! Hackerlar şifreni tahmin etmeye çalışmaz. Deauth saldırısıyla seni bir saniyede ağdan düşürürler. Telefonun tekrar bağlanırken şifreli paket havada yakalanır! Buna dört yollu el sıkışma denir. Hacker bu paketi güçlü bir bilgisayarda saniyeler içinde çözer. Çözüm çok basit: Halka açık ağlarda asla şifresiz gezme ve mutlaka VPN kullan! Siber dünyada av değil, avcı ol. Takipte kal!",
		Scenes:    6,
	},
	"2": {
		ID:        "2",
		Title:     "Şifreniz Ne Kadar Sürede Kırılır? (Brute Force Gerçeği)",
		Slug:      "sifren_nasil_kirilir_brute_force",
		Narration: "Kullandığın sekiz haneli şifre saniyeler içinde kırılabilir! Çoğu insan hala doğum tarihi veya bir iki üç dört beş altı kullanıyor. Hackerlar saniyede milyarlarca kombinasyon deneyen sistemler kullanır. Buna kaba kuvvet saldırısı denir. Sadece küçük harfli bir şifre anında çözülür. Ama içine büyük harf, rakam ve özel karakter eklersen? Kırılması yüzyıllar sürer! Şifreni hemen güncelle, parolalarını bir yöneticiye emanet et. Siber güvenlik şakaya gelmez!",
		Scenes:    6,
	},
	"3": {
		ID:        "3",
		Title:     "SQL Injection Nedir ve Nasıl Çalışır?",
		Slug:      "sql_injection_nasil_calisir",
		Narration: "Tek bir tırnak işaretiyle koca bir şirketin veritabanı boşaltılabilir mi? Evet! Buna SQL Injection denir. Web sitesine giriş yaparken kullanıcı adına 'veya bir eşittir bir' yazarsan ne olur? Sistem bunu kod zanneder ve şifre kontrolünü tamamen atlar! Hacker tek tıkla admin yetkisini ele geçirir, tüm kullanıcı verilerini çalar. Yazılımcılar dikkat: Kullanıcıdan gelen hiçbir veriye güvenmeyin, parametrik sorgular kullanın. Kodunu koru, geleceğini koru!",
		Scenes:    6,
	},
	"4": {
		ID:        "4",
		Title:     "Oltalama (Phishing) Tuzağına Nasıl Düşürülürsünüz?",
		Slug:      "phishing_oltalama_tuzagi",
		Narration: "Hesabınız askıya alındı, hemen tıklayın mesajı aldınız mı? Sakın tıklamayın! Bu bir oltalama tuzağıdır. Hackerlar bankanın veya popüler sitelerin birebir kopyasını yapar. Panikle linke tıklar ve şifrenizi girersiniz. Tebrikler, hesabınızı kendi ellerinizle teslim ettiniz! Gerçek kurumlar sizden asla mesajla şifre istemez. Şüpheli linklere tıklama, gönderici adresini mutlaka kontrol et. Tuzağa düşme, uyanık ol!",
		Scenes:    6,
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

