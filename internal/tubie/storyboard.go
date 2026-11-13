package tubie

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// StickmanActor represents a single stickman character in a scene.
type StickmanActor struct {
	Pose   int     `json:"pose"`
	X      int     `json:"x"`
	Y      int     `json:"y"`
	W      int     `json:"w"`
	Action string  `json:"action"` // "bounce", "shake", "panic", "jump", "run", "lean", "punch", "peek"
	From   float64 `json:"from,omitempty"`
	To     float64 `json:"to,omitempty"`
}

// SceneProp represents an icon or hardware asset placed in a scene.
type SceneProp struct {
	Path string  `json:"path"`
	X    int     `json:"x"`
	Y    int     `json:"y"`
	W    int     `json:"w"`
	From float64 `json:"from,omitempty"`
	To   float64 `json:"to,omitempty"`
}

// SceneDefinition is the high-level specification of a single video scene.
type SceneDefinition struct {
	Title     string          `json:"title"`
	Color     string          `json:"color"`
	Underline bool            `json:"underline"`
	SFX       string          `json:"sfx"`
	Stickmen  []StickmanActor `json:"stickmen,omitempty"`
	Prop      *SceneProp      `json:"prop,omitempty"`
	Arrow     bool            `json:"arrow"`
}

// BuildStoryboard compiles high-level scene definitions into rich autopilot shots with micro-acting and cyber staging.
func BuildStoryboard(scenes []SceneDefinition, totalDuration float64) []map[string]any {
	if len(scenes) == 0 {
		return []map[string]any{}
	}

	numScenes := len(scenes)
	sceneDur := totalDuration / float64(numScenes)
	shots := make([]map[string]any, 0, numScenes)

	availableProps := []string{
		"@b1/cyber-laptop.png",
		"@b1/cyber-skull.png",
		"@b1/cyber-shield.png",
		"@b1/cyber-terminal.png",
		"@b1/cyber-wifi.png",
		"@b1/cyber-padlock-unlocked.png",
		"@b1/cyber-padlock-locked.png",
		"@b1/idea-bulb-crayon.png",
	}

	currStart := 0.0
	for idx, sc := range scenes {
		start := currStart
		end := math.Round(float64(idx+1)*sceneDur*1000) / 1000
		if idx == numScenes-1 || end > totalDuration {
			end = totalDuration
		}
		if end <= start {
			end = start + 0.5
		}
		currStart = end
		rawSpan := end - start
		shotLen := math.Floor(rawSpan*1000) / 1000
		if shotLen <= 0.05 {
			shotLen = rawSpan
		}
		tSwitch := math.Floor(math.Min(shotLen*0.42, 1.8)*1000) / 1000

		titleText := strings.ToUpper(strings.TrimSpace(sc.Title))
		if titleText == "" {
			titleText = fmt.Sprintf("BÖLÜM 0%d", idx+1)
		}

		titleColor := sc.Color
		if titleColor == "" {
			colors := []string{"#00F0FF", "#FF3B30", "#FFCC00", "#00FF66", "#E056FD"}
			titleColor = colors[idx%len(colors)]
		}

		layers := make([]map[string]any, 0)

		// 1. Cyber Security HUD Header Badge
		hudText, hudColor := getHUDBadge(idx, titleText)
		layers = append(layers, map[string]any{
			"type":  "text",
			"text":  hudText,
			"x":     80,
			"y":     110,
			"w":     1120,
			"size":  30,
			"bold":  true,
			"color": hudColor,
			"keys": []map[string]any{
				{"t": 0.0, "alpha": 0.0},
				{"t": math.Round(math.Min(0.15, shotLen*0.3)*1000) / 1000, "alpha": 1.0},
			},
		})

		// 2. Main Scene Title Text
		tFade := math.Min(0.2, shotLen*0.4)
		layers = append(layers, map[string]any{
			"type":  "text",
			"text":  titleText,
			"x":     80,
			"y":     185,
			"w":     1120,
			"size":  62,
			"bold":  true,
			"color": titleColor,
			"keys": []map[string]any{
				{"t": 0.0, "alpha": 0.0},
				{"t": math.Round(tFade*1000) / 1000, "alpha": 1.0},
			},
		})

		// 3. Neon Underline Accent
		if sc.Underline || idx%2 == 0 {
			lineW := len(titleText) * 34
			if lineW < 380 {
				lineW = 380
			}
			if lineW > 1020 {
				lineW = 1020
			}
			tUnderline := math.Min(0.38, shotLen*0.75)
			layers = append(layers, map[string]any{
				"type":      "underline",
				"x":         80,
				"y":         275,
				"w":         lineW,
				"color":     titleColor,
				"lineWidth": 8,
				"keys": []map[string]any{
					{"t": 0.05, "progress": 0.0},
					{"t": math.Round(tUnderline*1000) / 1000, "progress": 1.0},
				},
			})
		}

		// 4. Dynamic Cyber Status Pill Badge
		pillText, pillColor := getStatusPill(idx, titleText)
		tPill1 := math.Min(0.18, shotLen*0.32)
		tPill2 := math.Min(0.32, shotLen*0.60)
		layers = append(layers, map[string]any{
			"type":   "text",
			"text":   pillText,
			"x":      640,
			"y":      340,
			"w":      960,
			"size":   24,
			"bold":   true,
			"color":  pillColor,
			"badge":  pillColor,
			"center": true,
			"keys": []map[string]any{
				{"t": 0.0, "alpha": 0.0, "y": 355},
				{"t": math.Round(tPill1*1000) / 1000, "alpha": 1.0, "y": 335},
				{"t": math.Round(tPill2*1000) / 1000, "alpha": 1.0, "y": 340},
			},
		})

		// 5. Stickman Actors with Mid-Scene Micro-Acting and Aura Lighting
		if len(sc.Stickmen) > 0 {
			for sIdx, sm := range sc.Stickmen {
				baseX := sm.X
				if baseX == 0 {
					baseX = 140 + sIdx*500
				}
				baseY := sm.Y
				if baseY == 0 {
					baseY = 640
				}
				baseW := sm.W
				if baseW == 0 {
					baseW = 500
				}
				poseIdx := (sm.Pose%3 + 3) % 3
				tStart := sm.From
				span := shotLen - tStart
				if sm.To > 0 && sm.To > sm.From {
					span = sm.To - sm.From
				}

				keys := buildActorKeyframes(sm.Action, baseX, baseY, baseW, tStart, span)
				layer := map[string]any{
					"type": "pose",
					"pose": poseIdx,
					"x":    baseX,
					"y":    baseY,
					"w":    baseW,
					"aura": "cyan",
					"keys": keys,
				}
				if sm.From > 0 {
					layer["from"] = sm.From
				}
				if sm.To > 0 {
					layer["to"] = sm.To
				}
				layers = append(layers, layer)
			}
		} else {
			// Automated high-engagement micro-acting based on narrative beats
			switch idx {
			case 0:
				// Scene 1: Victim browsing casually, then startled jump as hacker lurks
				// Victim Phase 1 (0 -> tSwitch): Suspicious browsing (pose 2)
				vKeys1 := buildActorKeyframes("lean", 140, 640, 500, 0.0, tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 2,
					"x":    140,
					"y":    640,
					"w":    500,
					"aura": "cyan",
					"to":   tSwitch,
					"keys": vKeys1,
				})
				// Victim Phase 2 (tSwitch -> shotLen): Shocked panic scream (pose 1)!
				vKeys2 := buildActorKeyframes("jump", 140, 640, 520, tSwitch, shotLen-tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 1,
					"x":    140,
					"y":    640,
					"w":    520,
					"aura": "cyan",
					"from": tSwitch,
					"keys": vKeys2,
				})
				// Alarm reaction exclamation mark above victim
				if tSwitch+0.08 < shotLen {
					tEx1 := math.Round(math.Min(tSwitch+0.12, shotLen*0.80)*1000) / 1000
					tEx2 := math.Round(math.Min(tSwitch+0.25, shotLen)*1000) / 1000
					exKeys := []map[string]any{
						{"t": tSwitch, "alpha": 0.0, "y": 595},
					}
					if tEx1 > tSwitch {
						exKeys = append(exKeys, map[string]any{"t": tEx1, "alpha": 1.0, "y": 555})
					}
					if tEx2 > tEx1 && tEx2 > tSwitch {
						exKeys = append(exKeys, map[string]any{"t": tEx2, "alpha": 1.0, "y": 570})
					}
					layers = append(layers, map[string]any{
						"type":  "text",
						"text":  "!!",
						"x":     320,
						"y":     570,
						"w":     200,
						"size":  64,
						"bold":  true,
						"color": "#FFE500",
						"from":  tSwitch,
						"keys":  exKeys,
					})
				}
				// Lurking Hacker (tSwitch -> shotLen): Peeking from right (pose 0)
				hKeys := buildActorKeyframes("peek", 640, 620, 500, tSwitch, shotLen-tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 0,
					"x":    640,
					"y":    620,
					"w":    500,
					"aura": "red",
					"from": tSwitch,
					"keys": hKeys,
				})

			case 1:
				// Scene 2: Victim clicking fake link, then full panic as trap springs
				// Victim Phase 1: Leaning into screen typing
				vKeys1 := buildActorKeyframes("lean", 140, 640, 500, 0.0, tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 0,
					"x":    140,
					"y":    640,
					"w":    500,
					"aura": "cyan",
					"to":   tSwitch,
					"keys": vKeys1,
				})
				// Victim Phase 2: Trap sprung, panic shake (pose 1)!
				vKeys2 := buildActorKeyframes("panic", 140, 640, 520, tSwitch, shotLen-tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 1,
					"x":    140,
					"y":    640,
					"w":    520,
					"aura": "cyan",
					"from": tSwitch,
					"keys": vKeys2,
				})
				// Panic exclamation
				if tSwitch+0.08 < shotLen {
					tEx1 := math.Round(math.Min(tSwitch+0.12, shotLen*0.80)*1000) / 1000
					tEx2 := math.Round(math.Min(tSwitch+0.25, shotLen)*1000) / 1000
					exKeys := []map[string]any{
						{"t": tSwitch, "alpha": 0.0, "y": 595},
					}
					if tEx1 > tSwitch {
						exKeys = append(exKeys, map[string]any{"t": tEx1, "alpha": 1.0, "y": 555})
					}
					if tEx2 > tEx1 && tEx2 > tSwitch {
						exKeys = append(exKeys, map[string]any{"t": tEx2, "alpha": 1.0, "y": 570})
					}
					layers = append(layers, map[string]any{
						"type":  "text",
						"text":  "!!",
						"x":     320,
						"y":     570,
						"w":     200,
						"size":  64,
						"bold":  true,
						"color": "#FF3B30",
						"from":  tSwitch,
						"keys":  exKeys,
					})
				}
				// Hacker celebrating right side
				hKeys := buildActorKeyframes("bounce", 640, 620, 500, tSwitch, shotLen-tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 0,
					"x":    640,
					"y":    620,
					"w":    500,
					"aura": "red",
					"from": tSwitch,
					"keys": hKeys,
				})

			case 2:
				// Scene 3: Chaos breach! Victim running, hacker punching keyboard
				vKeys := buildActorKeyframes("run", 140, 640, 520, 0.0, shotLen)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 1,
					"x":    140,
					"y":    640,
					"w":    520,
					"aura": "cyan",
					"keys": vKeys,
				})
				hKeys := buildActorKeyframes("punch", 640, 620, 520, 0.0, shotLen)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 0,
					"x":    640,
					"y":    620,
					"w":    520,
					"aura": "red",
					"keys": hKeys,
				})

			case 3:
				// Scene 4: Reality check / insight - Victim confused, then has epiphany
				vKeys1 := buildActorKeyframes("shake", 380, 640, 520, 0.0, tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 2,
					"x":    380,
					"y":    640,
					"w":    520,
					"aura": "cyan",
					"to":   tSwitch,
					"keys": vKeys1,
				})
				vKeys2 := buildActorKeyframes("bounce", 380, 640, 520, tSwitch, shotLen-tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 0,
					"x":    380,
					"y":    640,
					"w":    520,
					"aura": "cyan",
					"from": tSwitch,
					"keys": vKeys2,
				})

			case 4:
				// Scene 5: Defense 2FA - User prepares, then raises the shield in hero pose
				vKeys1 := buildActorKeyframes("lean", 380, 640, 520, 0.0, tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 2,
					"x":    380,
					"y":    640,
					"w":    520,
					"aura": "green",
					"to":   tSwitch,
					"keys": vKeys1,
				})
				vKeys2 := buildActorKeyframes("punch", 380, 620, 540, tSwitch, shotLen-tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 0,
					"x":    380,
					"y":    620,
					"w":    540,
					"aura": "green",
					"from": tSwitch,
					"keys": vKeys2,
				})

			default:
				// Scene 6: Victory & Viral CTA - Dual celebrating stickmen
				cKeys1 := buildActorKeyframes("bounce", 160, 640, 500, 0.0, shotLen)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 0,
					"x":    160,
					"y":    640,
					"w":    500,
					"aura": "cyan",
					"keys": cKeys1,
				})
				cKeys2 := buildActorKeyframes("jump", 620, 640, 500, 0.0, shotLen)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 0,
					"x":    620,
					"y":    640,
					"w":    500,
					"aura": "green",
					"keys": cKeys2,
				})
			}
		}

		// 6. Multi-Prop Staging (Aligned directly with characters & glowing + breathing)
		prop := sc.Prop
		if prop == nil {
			prop = &SceneProp{
				Path: availableProps[idx%len(availableProps)],
				X:    390,
				Y:    1060,
				W:    500,
			}
		}
		tProp1 := math.Round(math.Min(0.20, shotLen*0.35)*1000) / 1000
		tProp2 := math.Round(math.Min(0.35, shotLen*0.65)*1000) / 1000
		propGlow := "cyan"
		if idx == 1 || idx == 2 {
			propGlow = "red"
		} else if idx == 3 {
			propGlow = "yellow"
		} else if idx == 4 {
			propGlow = "green"
		}
		propKeys := []map[string]any{
			{"t": 0.0, "w": 80, "alpha": 0.0, "angle": 0.0},
		}
		lastPropT := 0.0
		if tProp1 > lastPropT {
			propKeys = append(propKeys, map[string]any{"t": tProp1, "w": prop.W + 45, "alpha": 1.0, "angle": 0.05})
			lastPropT = tProp1
		}
		if tProp2 > lastPropT {
			propKeys = append(propKeys, map[string]any{"t": tProp2, "w": prop.W, "alpha": 1.0, "angle": 0.0})
			lastPropT = tProp2
		}
		// Subtle centered terminal hover/breathing pulse for the remainder of the shot
		pCur := lastPropT + 0.75
		pTog := false
		for pCur < shotLen-0.15 {
			rp := math.Round(pCur*1000) / 1000
			if rp > lastPropT {
				pw := prop.W + 12
				px := prop.X - 6
				py := prop.Y - 6
				if pTog {
					pw = prop.W
					px = prop.X
					py = prop.Y
				}
				propKeys = append(propKeys, map[string]any{
					"t": rp,
					"x": px,
					"y": py,
					"w": pw,
				})
				lastPropT = rp
			}
			pTog = !pTog
			pCur += 0.75
		}
		layers = append(layers, map[string]any{
			"type": "image",
			"path": prop.Path,
			"x":    prop.X,
			"y":    prop.Y,
			"w":    prop.W,
			"glow": propGlow,
			"keys": propKeys,
		})

		// Secondary Prop (Staged support asset with glow & zero-G hover)
		secProp := getSecondaryProp(idx, prop.Path)
		if secProp != nil && tSwitch+0.25 < shotLen {
			secT1 := math.Round(math.Min(tSwitch+0.16, shotLen*0.75)*1000) / 1000
			secT2 := math.Round(math.Min(tSwitch+0.32, shotLen*0.92)*1000) / 1000
			secGlow := "cyan"
			if idx == 1 || idx == 2 {
				secGlow = "red"
			} else if idx == 4 {
				secGlow = "green"
			}
			secKeys := []map[string]any{
				{"t": tSwitch, "w": 50, "alpha": 0.0, "angle": 0.0},
			}
			lastSecT := tSwitch
			if secT1 > lastSecT {
				secKeys = append(secKeys, map[string]any{"t": secT1, "w": secProp.W + 30, "alpha": 1.0, "angle": -0.06})
				lastSecT = secT1
			}
			if secT2 > lastSecT {
				secKeys = append(secKeys, map[string]any{"t": secT2, "w": secProp.W, "alpha": 1.0, "angle": 0.0})
				lastSecT = secT2
			}
			sCur := lastSecT + 0.65
			sTog := false
			for sCur < shotLen-0.15 {
				rs := math.Round(sCur*1000) / 1000
				if rs > lastSecT {
					sy := secProp.Y - 12
					sAng := 0.04
					if sTog {
						sy = secProp.Y + 4
						sAng = -0.04
					}
					secKeys = append(secKeys, map[string]any{
						"t":     rs,
						"y":     sy,
						"angle": sAng,
					})
					lastSecT = rs
				}
				sTog = !sTog
				sCur += 0.65
			}
			secLayer := map[string]any{
				"type": "image",
				"path": secProp.Path,
				"x":    secProp.X,
				"y":    secProp.Y,
				"w":    secProp.W,
				"glow": secGlow,
				"from": tSwitch,
				"keys": secKeys,
			}
			layers = append(layers, secLayer)
		}

		// 7. Threat Direction Arrow (Arcing cleanly above actors and !! alert)
		if sc.Arrow || idx == 1 || idx == 2 {
			tArrow := math.Min(0.42, shotLen*0.65)
			layers = append(layers, map[string]any{
				"type":      "arrow",
				"x":         340,
				"y":         510,
				"w":         600,
				"color":     "#FF3B30",
				"lineWidth": 5,
				"keys": []map[string]any{
					{"t": 0.0, "progress": 0.0},
					{"t": math.Round(tArrow*1000) / 1000, "progress": 1.0},
				},
			})
		}

		shots = append(shots, map[string]any{
			"start":  start,
			"end":    end,
			"layers": layers,
		})
	}

	return shots
}

func getHUDBadge(idx int, title string) (string, string) {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "tuzak") || strings.Contains(lower, "sahte") || strings.Contains(lower, "dm"):
		return "[ ALERT // OLTALAMA SALDIRISI ]", "#FF3B30"
	case strings.Contains(lower, "ele geçiril") || strings.Contains(lower, "çalındı") || strings.Contains(lower, "şifre"):
		return "[ SİSTEM İHLALİ // HESAP ELE GEÇİRİLDİ ]", "#FF3B30"
	case strings.Contains(lower, "kaybedil") || strings.Contains(lower, "kontrol") || strings.Contains(lower, "sızma"):
		return "[ KRİTİK ALARM // KONTROL KAYBI ]", "#FFCC00"
	case strings.Contains(lower, "istemez") || strings.Contains(lower, "gerçek") || strings.Contains(lower, "farkındalık") || strings.Contains(lower, "halka"):
		return "[ SİBER GERÇEK // TUZAĞA DÜŞME ]", "#E056FD"
	case strings.Contains(lower, "koru") || strings.Contains(lower, "2fa") || strings.Contains(lower, "doğrulama") || strings.Contains(lower, "kalkan"):
		return "[ GÜVENLİK DUVARI // 2FA SAVUNMASI ]", "#00FF66"
	case strings.Contains(lower, "uyanık") || strings.Contains(lower, "güvende") || strings.Contains(lower, "takip") || strings.Contains(lower, "avcı"):
		return "[ KORUMA AKTİF // GÜVENDESİN ]", "#00F0FF"
	default:
		hudList := []struct {
			text  string
			color string
		}{
			{"[ SİBER İSTİHBARAT // ACİL UYARI ]", "#00F0FF"},
			{"[ TEHLİKE DÜZEYİ // KRİTİK ALARM ]", "#FF3B30"},
			{"[ SİSTEM İHLALİ // RİSK TESPİTİ ]", "#FFCC00"},
			{"[ GÜVENLİK GERÇEĞİ // DİKKAT ]", "#E056FD"},
			{"[ AKTİF SAVUNMA // SİBER KALKAN ]", "#00FF66"},
			{"[ GÜVENDE KAL // BİLİNÇLİ AVCI ]", "#00F0FF"},
		}
		item := hudList[idx%len(hudList)]
		return item.text, item.color
	}
}

func getStatusPill(idx int, title string) (string, string) {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "tuzak") || strings.Contains(lower, "sahte") || strings.Contains(lower, "dm"):
		return "[ TEHLİKELİ BAĞLANTI TESPİT EDİLDİ ]", "#FF3B30"
	case strings.Contains(lower, "ele geçiril") || strings.Contains(lower, "çalındı") || strings.Contains(lower, "şifre"):
		return "[ ALARM: KİMLİK BİLGİLERİ SIZDIRILDI ]", "#FF3B30"
	case strings.Contains(lower, "kaybedil") || strings.Contains(lower, "kontrol") || strings.Contains(lower, "sızma"):
		return "[ KRİTİK İHLAL: HESAP KONTROLÜ KAYIP ]", "#FFCC00"
	case strings.Contains(lower, "istemez") || strings.Contains(lower, "gerçek") || strings.Contains(lower, "farkındalık") || strings.Contains(lower, "halka"):
		return "[ SİBER GERÇEK: EN ZAYIF HALKA İNSAN ]", "#E056FD"
	case strings.Contains(lower, "koru") || strings.Contains(lower, "2fa") || strings.Contains(lower, "doğrulama") || strings.Contains(lower, "kalkan"):
		return "[ 2FA DOĞRULAMA: GÜVENLİK KALKANI DEVREDE ]", "#00FF66"
	case strings.Contains(lower, "uyanık") || strings.Contains(lower, "güvende") || strings.Contains(lower, "takip") || strings.Contains(lower, "avcı"):
		return "[ TAM SAVUNMA: AV DEĞİL AVCI OL ]", "#00F0FF"
	default:
		pills := []struct {
			text  string
			color string
		}{
			{"[ OLTALAMA SALDIRISI TESPİT EDİLDİ ]", "#00F0FF"},
			{"[ TEHLİKELİ BAĞLANTI UYARISI ]", "#FF3B30"},
			{"[ ALARM: ŞİFRE SIZDIRILDI ]", "#FFCC00"},
			{"[ DİKKAT: ŞÜPHELİ LİNKLERİ SİL ]", "#E056FD"},
			{"[ 2FA AKTİF: HESAP GÜVENDE ]", "#00FF66"},
			{"[ BİLİNÇLİ KULLANICI // TAM KORUMA ]", "#00F0FF"},
		}
		item := pills[idx%len(pills)]
		return item.text, item.color
	}
}

func getSecondaryProp(idx int, primaryPath string) *SceneProp {
	var prop *SceneProp
	switch idx {
	case 0:
		prop = &SceneProp{Path: "@b1/cyber-wifi.png", X: 960, Y: 1135, W: 230}
	case 1:
		prop = &SceneProp{Path: "@b1/cyber-padlock-unlocked.png", X: 90, Y: 1090, W: 240}
	case 2:
		prop = &SceneProp{Path: "@b1/cyber-bug.png", X: 90, Y: 1090, W: 240}
	case 3:
		prop = &SceneProp{Path: "@b1/cyber-database.png", X: 960, Y: 1125, W: 230}
	case 4:
		prop = &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 90, Y: 1090, W: 240}
	case 5:
		prop = &SceneProp{Path: "@b1/cyber-shield.png", X: 960, Y: 1125, W: 240}
	default:
		prop = &SceneProp{Path: "@b1/idea-bulb-crayon.png", X: 960, Y: 1125, W: 230}
	}
	if prop.Path == primaryPath {
		if primaryPath != "@b1/cyber-skull.png" && idx < 3 {
			prop.Path = "@b1/cyber-skull.png"
		} else if primaryPath != "@b1/cyber-database.png" {
			prop.Path = "@b1/cyber-database.png"
		} else {
			prop.Path = "@b1/cyber-shield.png"
		}
	}
	return prop
}

func buildActorKeyframes(action string, baseX, baseY, baseW int, tStart, span float64) []map[string]any {
	t0 := math.Floor(tStart*1000) / 1000
	maxSpan := span
	if maxSpan <= 0.01 {
		maxSpan = 0.01
	}
	maxT := math.Floor((t0+maxSpan)*1000) / 1000

	keys := make([]map[string]any, 0)
	lastT := -1.0
	roundT := func(val float64) float64 {
		return math.Floor(val*1000) / 1000
	}
	addKey := func(tVal float64, kv map[string]any) {
		rt := roundT(tVal)
		if rt > maxT {
			rt = maxT
		}
		if rt <= lastT {
			return
		}
		kv["t"] = rt
		keys = append(keys, kv)
		lastT = rt
	}

	switch action {
	case "panic":
		// Initial violent shock jump and shudder
		addKey(t0, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": 0.0})
		addKey(t0+0.08, map[string]any{"x": baseX - 30, "y": baseY - 20, "w": baseW + 15, "angle": -0.14})
		addKey(t0+0.16, map[string]any{"x": baseX + 30, "y": baseY - 5, "w": baseW + 10, "angle": 0.14})
		addKey(t0+0.24, map[string]any{"x": baseX - 18, "y": baseY - 12, "w": baseW, "angle": -0.09})
		addKey(t0+0.32, map[string]any{"x": baseX + 18, "y": baseY, "w": baseW, "angle": 0.09})
		addKey(t0+0.40, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": 0.0})

		// Continuous high-tension nervous tremor throughout the scene
		step := 0.14
		cur := t0 + 0.40 + step
		toggle := false
		for cur < maxT-0.05 {
			jitterX := baseX + 8
			jitterAngle := 0.04
			if toggle {
				jitterX = baseX - 8
				jitterAngle = -0.04
			}
			addKey(cur, map[string]any{
				"x":     jitterX,
				"y":     baseY + int(math.Sin(cur*15.0)*5.0),
				"w":     baseW,
				"angle": jitterAngle,
			})
			toggle = !toggle
			cur += step
		}
		addKey(maxT, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": 0.0})

	case "shake":
		// Head-shaking in disbelief, then settling into agitated contemplation
		addKey(t0, map[string]any{"x": baseX, "y": baseY, "w": baseW})
		addKey(t0+0.08, map[string]any{"x": baseX - 16, "y": baseY, "w": baseW})
		addKey(t0+0.16, map[string]any{"x": baseX + 16, "y": baseY, "w": baseW})
		addKey(t0+0.24, map[string]any{"x": baseX - 10, "y": baseY, "w": baseW})
		addKey(t0+0.32, map[string]any{"x": baseX + 10, "y": baseY, "w": baseW})
		addKey(t0+0.40, map[string]any{"x": baseX, "y": baseY, "w": baseW})

		step := 0.40
		cur := t0 + 0.40 + step
		for cur < maxT-0.05 {
			addKey(cur, map[string]any{"x": baseX, "y": baseY - 12, "w": baseW + 6})
			if cur+step*0.5 < maxT-0.05 {
				addKey(cur+step*0.5, map[string]any{"x": baseX, "y": baseY, "w": baseW})
			}
			cur += step
		}
		addKey(maxT, map[string]any{"x": baseX, "y": baseY, "w": baseW})

	case "jump":
		// High-impact dramatic leap with squash and stretch, followed by dynamic posture
		t1 := math.Min(0.20, maxSpan*0.40)
		t2 := math.Min(0.36, maxSpan*0.70)
		addKey(t0, map[string]any{"x": baseX, "y": baseY + 180, "w": baseW - 20, "alpha": 0.0})
		addKey(t0+t1, map[string]any{"x": baseX, "y": baseY - 45, "w": baseW + 20, "alpha": 1.0})
		addKey(t0+t2, map[string]any{"x": baseX, "y": baseY, "w": baseW, "alpha": 1.0})

		step := 0.42
		cur := t0 + t2 + step
		for cur < maxT-0.05 {
			addKey(cur, map[string]any{"x": baseX, "y": baseY - 16, "w": baseW + 8})
			if cur+step*0.5 < maxT-0.05 {
				addKey(cur+step*0.5, map[string]any{"x": baseX, "y": baseY, "w": baseW})
			}
			cur += step
		}
		addKey(maxT, map[string]any{"x": baseX, "y": baseY, "w": baseW, "alpha": 1.0})

	case "run":
		// Sprints urgently into frame, keeping energetic movement
		t1 := math.Min(0.22, maxSpan*0.40)
		t2 := math.Min(0.42, maxSpan*0.75)
		addKey(t0, map[string]any{"x": baseX - 220, "y": baseY, "w": baseW, "angle": 0.14, "alpha": 0.0})
		addKey(t0+t1, map[string]any{"x": baseX - 60, "y": baseY - 28, "w": baseW, "angle": 0.09, "alpha": 1.0})
		addKey(t0+t2, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": 0.0, "alpha": 1.0})

		step := 0.35
		cur := t0 + t2 + step
		toggle := false
		for cur < maxT-0.05 {
			swayX := baseX + 14
			if toggle {
				swayX = baseX - 14
			}
			addKey(cur, map[string]any{
				"x":     swayX,
				"y":     baseY - 14,
				"w":     baseW,
				"angle": 0.04 * (map[bool]float64{true: -1, false: 1}[toggle]),
			})
			toggle = !toggle
			cur += step
		}
		addKey(maxT, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": 0.0, "alpha": 1.0})

	case "lean":
		// Active posture: leans forward into terminal, with continuous subtle typing pulses
		addKey(t0, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": 0.0})
		addKey(t0+0.16, map[string]any{"x": baseX + 18, "y": baseY + 12, "w": baseW, "angle": 0.12})

		step := 0.40
		cur := t0 + 0.16 + step
		toggle := false
		for cur < maxT-0.05 {
			leanAngle := 0.14
			leanY := baseY + 14
			if toggle {
				leanAngle = 0.10
				leanY = baseY + 10
			}
			addKey(cur, map[string]any{
				"x":     baseX + 18,
				"y":     leanY,
				"w":     baseW,
				"angle": leanAngle,
			})
			toggle = !toggle
			cur += step
		}
		addKey(maxT, map[string]any{"x": baseX + 18, "y": baseY + 12, "w": baseW, "angle": 0.12})

	case "punch":
		// High-energy hacker keystroke flurry pounding on keyboard
		addKey(t0, map[string]any{"x": baseX, "y": baseY, "w": baseW})
		addKey(t0+0.12, map[string]any{"x": baseX - 18, "y": baseY - 24, "w": baseW + 50})
		addKey(t0+0.24, map[string]any{"x": baseX, "y": baseY, "w": baseW})

		step := 0.22
		cur := t0 + 0.24 + step
		toggle := false
		for cur < maxT-0.05 {
			strikeY := baseY - 18
			strikeW := baseW + 35
			strikeX := baseX - 12
			if toggle {
				strikeY = baseY
				strikeW = baseW
				strikeX = baseX
			}
			addKey(cur, map[string]any{
				"x": strikeX,
				"y": strikeY,
				"w": strikeW,
			})
			toggle = !toggle
			cur += step
		}
		addKey(maxT, map[string]any{"x": baseX, "y": baseY, "w": baseW})

	case "peek":
		// Lurking hacker creeps in from right with stealth angle, watching victim
		t1 := math.Min(0.24, maxSpan*0.45)
		t2 := math.Min(0.48, maxSpan*0.80)
		addKey(t0, map[string]any{"x": baseX + 160, "y": baseY, "w": baseW, "angle": -0.10, "alpha": 0.0})
		addKey(t0+t1, map[string]any{"x": baseX + 40, "y": baseY - 12, "w": baseW, "angle": -0.06, "alpha": 1.0})
		addKey(t0+t2, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": -0.03, "alpha": 1.0})

		step := 0.45
		cur := t0 + t2 + step
		toggle := false
		for cur < maxT-0.05 {
			addKey(cur, map[string]any{
				"x":     baseX + map[bool]int{true: 12, false: -8}[toggle],
				"y":     baseY + map[bool]int{true: -10, false: 0}[toggle],
				"w":     baseW,
				"angle": -0.03,
			})
			toggle = !toggle
			cur += step
		}
		addKey(maxT, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": -0.03, "alpha": 1.0})

	default: // "bounce"
		step := 0.42
		cur := t0
		toggle := false
		addKey(t0, map[string]any{"x": baseX, "y": baseY, "w": baseW})
		cur += step * 0.5
		for cur < maxT-0.05 {
			bY := baseY - 28
			bW := baseW + 12
			if toggle {
				bY = baseY
				bW = baseW
			}
			addKey(cur, map[string]any{
				"x": baseX,
				"y": bY,
				"w": bW,
			})
			toggle = !toggle
			cur += step * 0.5
		}
		addKey(maxT, map[string]any{"x": baseX, "y": baseY, "w": baseW})
	}

	return keys
}

// BuildSoundDesign compiles rich, synchronized SFX cues for high-retention video pacing.
func BuildSoundDesign(scenes []SceneDefinition, totalDuration float64) []map[string]any {
	if len(scenes) == 0 || totalDuration <= 0.5 {
		return []map[string]any{}
	}

	numScenes := len(scenes)
	sceneDur := totalDuration / float64(numScenes)
	sfxList := make([]map[string]any, 0)

	for i, sc := range scenes {
		start := math.Round(float64(i)*sceneDur*1000) / 1000
		shotLen := sceneDur
		if i == numScenes-1 {
			shotLen = totalDuration - start
		}
		tSwitch := math.Round(math.Min(shotLen*0.42, 1.8)*1000) / 1000

		// 1. Scene Title Pop
		tPop := math.Round((start+0.04)*1000) / 1000
		if tPop < totalDuration-0.1 {
			sfxList = append(sfxList, map[string]any{
				"time": tPop,
				"kind": "pop",
				"gain": 0.18,
			})
		}

		// 2. Underline Swipe / Air whoosh
		if sc.Underline || i%2 == 0 {
			tSwipe := math.Round((start+0.18)*1000) / 1000
			if tSwipe < totalDuration-0.1 {
				sfxList = append(sfxList, map[string]any{
					"time": tSwipe,
					"kind": "paper",
					"gain": 0.14,
				})
			}
		}

		// 3. Scene Event Reaction SFX (Timed to character reaction & secondary prop drop)
		actionTime := math.Round((start+tSwitch)*1000) / 1000
		if actionTime < totalDuration-0.1 {
			kind := sc.SFX
			if kind == "" {
				switch i {
				case 0:
					kind = "keyboard"
				case 1:
					kind = "alert"
				case 2:
					kind = "glitch"
				case 3:
					kind = "chime"
				case 4:
					kind = "lock"
				default:
					kind = "jingle"
				}
			}

			switch kind {
			case "keyboard":
				for kIdx, offset := range []float64{0.0, 0.13, 0.26} {
					tKey := math.Round((actionTime+offset)*1000) / 1000
					if tKey < totalDuration-0.1 {
						sfxList = append(sfxList, map[string]any{
							"time": tKey,
							"kind": "keyboard",
							"gain": 0.18 - float64(kIdx)*0.02,
						})
					}
				}
			case "alert":
				sfxList = append(sfxList, map[string]any{
					"time": actionTime,
					"kind": "alert",
					"gain": 0.22,
				})
			case "glitch":
				tPre := math.Round((start+0.42)*1000) / 1000
				if tPre < actionTime && tPre < totalDuration-0.1 {
					sfxList = append(sfxList, map[string]any{
						"time": tPre,
						"kind": "keyboard",
						"gain": 0.16,
					})
				}
				sfxList = append(sfxList, map[string]any{
					"time": actionTime,
					"kind": "glitch",
					"gain": 0.20,
				})
			case "lock":
				sfxList = append(sfxList, map[string]any{
					"time": actionTime,
					"kind": "lock",
					"gain": 0.20,
				})
			case "chime":
				sfxList = append(sfxList, map[string]any{
					"time": actionTime,
					"kind": "chime",
					"gain": 0.18,
				})
			case "jingle":
				sfxList = append(sfxList, map[string]any{
					"time": actionTime,
					"kind": "jingle",
					"gain": 0.20,
				})
			default:
				sfxList = append(sfxList, map[string]any{
					"time": actionTime,
					"kind": "tap",
					"gain": 0.15,
				})
			}
		}

		// 4. Secondary prop drop / micro-reaction cue
		tSec := math.Round((start+tSwitch+0.18)*1000) / 1000
		if tSec < totalDuration-0.1 {
			sfxList = append(sfxList, map[string]any{
				"time": tSec,
				"kind": "tap",
				"gain": 0.14,
			})
		}
	}

	// 5. Climax / Victory resolution in final scene
	tClimax := math.Round((totalDuration-0.65)*1000) / 1000
	if tClimax > 0.5 {
		sfxList = append(sfxList, map[string]any{
			"time": tClimax,
			"kind": "jingle",
			"gain": 0.20,
		})
	}

	// Sort strictly by timestamp
	sort.Slice(sfxList, func(i, j int) bool {
		return sfxList[i]["time"].(float64) < sfxList[j]["time"].(float64)
	})

	return sfxList
}
