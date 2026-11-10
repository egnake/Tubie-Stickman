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
		shotLen := end - start
		tSwitch := math.Round(math.Min(shotLen*0.42, 1.8)*1000) / 1000

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
			"y":     130,
			"w":     1120,
			"size":  28,
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
			"y":     220,
			"w":     1120,
			"size":  58,
			"bold":  true,
			"color": titleColor,
			"keys": []map[string]any{
				{"t": 0.0, "alpha": 0.0},
				{"t": math.Round(tFade*1000) / 1000, "alpha": 1.0},
			},
		})

		// 3. Neon Underline Accent
		if sc.Underline || idx%2 == 0 {
			lineW := len(titleText) * 32
			if lineW < 360 {
				lineW = 360
			}
			if lineW > 1000 {
				lineW = 1000
			}
			tUnderline := math.Min(0.40, shotLen*0.75)
			layers = append(layers, map[string]any{
				"type":      "underline",
				"x":         80,
				"y":         310,
				"w":         lineW,
				"color":     titleColor,
				"lineWidth": 7,
				"keys": []map[string]any{
					{"t": 0.05, "progress": 0.0},
					{"t": math.Round(tUnderline*1000) / 1000, "progress": 1.0},
				},
			})
		}

		// 4. Stickman Actors with Mid-Scene Micro-Acting
		if len(sc.Stickmen) > 0 {
			for sIdx, sm := range sc.Stickmen {
				baseX := sm.X
				if baseX == 0 {
					baseX = 180 + sIdx*500
				}
				baseY := sm.Y
				if baseY == 0 {
					baseY = 600
				}
				baseW := sm.W
				if baseW == 0 {
					baseW = 460
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
				// Victim Phase 1 (0 -> tSwitch): Curious browsing
				vKeys1 := buildActorKeyframes("lean", 220, 600, 450, 0.0, tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 0,
					"x":    220,
					"y":    600,
					"w":    450,
					"to":   tSwitch,
					"keys": vKeys1,
				})
				// Victim Phase 2 (tSwitch -> shotLen): Shocked jump!
				vKeys2 := buildActorKeyframes("jump", 220, 600, 470, tSwitch, shotLen-tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 2,
					"x":    220,
					"y":    600,
					"w":    470,
					"from": tSwitch,
					"keys": vKeys2,
				})
				// Lurking Hacker (tSwitch -> shotLen): Peeking from right
				hKeys := buildActorKeyframes("peek", 680, 580, 460, tSwitch, shotLen-tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 1,
					"x":    680,
					"y":    580,
					"w":    460,
					"from": tSwitch,
					"keys": hKeys,
				})

			case 1:
				// Scene 2: Victim clicking fake link, then full panic as trap springs
				// Victim Phase 1: Leaning into screen
				vKeys1 := buildActorKeyframes("lean", 240, 600, 450, 0.0, tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 0,
					"x":    240,
					"y":    600,
					"w":    450,
					"to":   tSwitch,
					"keys": vKeys1,
				})
				// Victim Phase 2: Trap sprung, panic shake!
				vKeys2 := buildActorKeyframes("panic", 240, 600, 470, tSwitch, shotLen-tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 2,
					"x":    240,
					"y":    600,
					"w":    470,
					"from": tSwitch,
					"keys": vKeys2,
				})
				// Hacker celebrating right side
				hKeys := buildActorKeyframes("bounce", 680, 580, 460, tSwitch, shotLen-tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 1,
					"x":    680,
					"y":    580,
					"w":    460,
					"from": tSwitch,
					"keys": hKeys,
				})

			case 2:
				// Scene 3: Chaos breach! Victim running, hacker punching keyboard
				vKeys := buildActorKeyframes("run", 220, 600, 480, 0.0, shotLen)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 2,
					"x":    220,
					"y":    600,
					"w":    480,
					"keys": vKeys,
				})
				hKeys := buildActorKeyframes("punch", 680, 580, 480, 0.0, shotLen)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 1,
					"x":    680,
					"y":    580,
					"w":    480,
					"keys": hKeys,
				})

			case 3:
				// Scene 4: Reality check / insight - Victim confused, then has epiphany
				vKeys1 := buildActorKeyframes("shake", 420, 600, 460, 0.0, tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 2,
					"x":    420,
					"y":    600,
					"w":    460,
					"to":   tSwitch,
					"keys": vKeys1,
				})
				vKeys2 := buildActorKeyframes("bounce", 420, 600, 460, tSwitch, shotLen-tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 0,
					"x":    420,
					"y":    600,
					"w":    460,
					"from": tSwitch,
					"keys": vKeys2,
				})

			case 4:
				// Scene 5: Defense 2FA - User prepares, then raises the shield in hero pose
				vKeys1 := buildActorKeyframes("lean", 420, 600, 460, 0.0, tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 0,
					"x":    420,
					"y":    600,
					"w":    460,
					"to":   tSwitch,
					"keys": vKeys1,
				})
				vKeys2 := buildActorKeyframes("punch", 420, 580, 490, tSwitch, shotLen-tSwitch)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 1,
					"x":    420,
					"y":    580,
					"w":    490,
					"from": tSwitch,
					"keys": vKeys2,
				})

			default:
				// Scene 6: Victory & Viral CTA - Dual celebrating stickmen
				cKeys1 := buildActorKeyframes("bounce", 240, 600, 460, 0.0, shotLen)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 1,
					"x":    240,
					"y":    600,
					"w":    460,
					"keys": cKeys1,
				})
				cKeys2 := buildActorKeyframes("jump", 680, 600, 460, 0.0, shotLen)
				layers = append(layers, map[string]any{
					"type": "pose",
					"pose": 0,
					"x":    680,
					"y":    600,
					"w":    460,
					"keys": cKeys2,
				})
			}
		}

		// 5. Multi-Prop Staging (Primary + Secondary Indicators)
		prop := sc.Prop
		if prop == nil {
			prop = &SceneProp{
				Path: availableProps[idx%len(availableProps)],
				X:    490,
				Y:    1220,
				W:    300,
			}
		}
		tProp1 := math.Min(0.22, shotLen*0.4)
		tProp2 := math.Min(0.38, shotLen*0.7)
		layers = append(layers, map[string]any{
			"type": "image",
			"path": prop.Path,
			"x":    prop.X,
			"y":    prop.Y,
			"w":    prop.W,
			"keys": []map[string]any{
				{"t": 0.0, "w": 60, "alpha": 0.0, "angle": 0.0},
				{"t": math.Round(tProp1*1000) / 1000, "w": prop.W + 35, "alpha": 1.0, "angle": 0.06},
				{"t": math.Round(tProp2*1000) / 1000, "w": prop.W, "alpha": 1.0, "angle": 0.0},
			},
		})

		// Secondary Prop (Staged support asset)
		secProp := getSecondaryProp(idx, prop.Path)
		if secProp != nil {
			secT1 := math.Round(math.Min(tSwitch+0.18, shotLen*0.8)*1000) / 1000
			secT2 := math.Round(math.Min(tSwitch+0.34, shotLen*0.95)*1000) / 1000
			secLayer := map[string]any{
				"type": "image",
				"path": secProp.Path,
				"x":    secProp.X,
				"y":    secProp.Y,
				"w":    secProp.W,
				"from": tSwitch,
				"keys": []map[string]any{
					{"t": tSwitch, "w": 40, "alpha": 0.0, "angle": 0.0},
					{"t": secT1, "w": secProp.W + 25, "alpha": 1.0, "angle": -0.06},
					{"t": secT2, "w": secProp.W, "alpha": 1.0, "angle": 0.0},
				},
			}
			layers = append(layers, secLayer)
		}

		// 6. Threat Direction Arrow
		if sc.Arrow || idx == 1 || idx == 2 {
			tArrow := math.Min(0.45, shotLen*0.7)
			layers = append(layers, map[string]any{
				"type":  "arrow",
				"x":     360,
				"y":     1320,
				"w":     540,
				"color": "#FF3B30",
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
		return "[ ⚠️ TEHLİKE DÜZEYİ: OLTALAMA SALDIRISI ]", "#FF3B30"
	case strings.Contains(lower, "ele geçiril") || strings.Contains(lower, "çalındı") || strings.Contains(lower, "şifre"):
		return "[ 🚨 SİSTEM İHLALİ: HESAP ELE GEÇİRİLDİ ]", "#FF3B30"
	case strings.Contains(lower, "kaybedil") || strings.Contains(lower, "kontrol") || strings.Contains(lower, "sızma"):
		return "[ ⚠️ KRİTİK ALARM: KONTROL KAYBI ]", "#FFCC00"
	case strings.Contains(lower, "istemez") || strings.Contains(lower, "gerçek") || strings.Contains(lower, "farkındalık") || strings.Contains(lower, "halka"):
		return "[ 💡 SİBER GERÇEK // TUZAĞA DÜŞME ]", "#E056FD"
	case strings.Contains(lower, "koru") || strings.Contains(lower, "2fa") || strings.Contains(lower, "doğrulama") || strings.Contains(lower, "kalkan"):
		return "[ 🛡️ GÜVENLİK DUVARI: İKİ ADIMLI KORUMA ]", "#00FF66"
	case strings.Contains(lower, "uyanık") || strings.Contains(lower, "güvende") || strings.Contains(lower, "takip") || strings.Contains(lower, "avcı"):
		return "[ 🏆 KORUMA AKTİF // GÜVENDESİN ]", "#00F0FF"
	default:
		hudList := []struct {
			text  string
			color string
		}{
			{"[ ⚡ SİBER İSTİHBARAT // ACİL UYARI ]", "#00F0FF"},
			{"[ ⚠️ TEHLİKE DÜZEYİ: KRİTİK ALARM ]", "#FF3B30"},
			{"[ 🚨 SİSTEM İHLALİ: RİSK TESPİTİ ]", "#FFCC00"},
			{"[ 💡 GÜVENLİK GERÇEĞİ: DİKKAT ]", "#E056FD"},
			{"[ 🛡️ GÜVENLİK DUVARI: AKTİF SAVUNMA ]", "#00FF66"},
			{"[ 🏆 GÜVENDE KAL // TAKİP ET ]", "#00F0FF"},
		}
		item := hudList[idx%len(hudList)]
		return item.text, item.color
	}
}

func getSecondaryProp(idx int, primaryPath string) *SceneProp {
	switch idx {
	case 0:
		return &SceneProp{Path: "@b1/cyber-wifi.png", X: 800, Y: 1240, W: 190}
	case 1:
		return &SceneProp{Path: "@b1/cyber-skull.png", X: 180, Y: 1220, W: 210}
	case 2:
		return &SceneProp{Path: "@b1/cyber-bug.png", X: 180, Y: 1260, W: 190}
	case 3:
		return &SceneProp{Path: "@b1/cyber-padlock-unlocked.png", X: 800, Y: 1240, W: 190}
	case 4:
		return &SceneProp{Path: "@b1/cyber-padlock-locked.png", X: 180, Y: 1240, W: 200}
	case 5:
		return &SceneProp{Path: "@b1/cyber-shield.png", X: 800, Y: 1240, W: 200}
	default:
		if primaryPath != "@b1/idea-bulb-crayon.png" {
			return &SceneProp{Path: "@b1/idea-bulb-crayon.png", X: 800, Y: 1240, W: 190}
		}
		return &SceneProp{Path: "@b1/cyber-shield.png", X: 180, Y: 1240, W: 200}
	}
}

func buildActorKeyframes(action string, baseX, baseY, baseW int, tStart, span float64) []map[string]any {
	t0 := math.Round(tStart*1000) / 1000
	if span < 0.2 {
		span = 0.2
	}
	switch action {
	case "shake":
		tEnd := math.Min(0.40, span*0.9)
		tStep := tEnd / 5.0
		return []map[string]any{
			{"t": t0, "x": baseX, "y": baseY, "w": baseW},
			{"t": math.Round((t0+tStep*1)*1000) / 1000, "x": baseX - 14, "y": baseY, "w": baseW},
			{"t": math.Round((t0+tStep*2)*1000) / 1000, "x": baseX + 14, "y": baseY, "w": baseW},
			{"t": math.Round((t0+tStep*3)*1000) / 1000, "x": baseX - 8, "y": baseY, "w": baseW},
			{"t": math.Round((t0+tStep*4)*1000) / 1000, "x": baseX + 8, "y": baseY, "w": baseW},
			{"t": math.Round((t0+tStep*5)*1000) / 1000, "x": baseX, "y": baseY, "w": baseW},
		}

	case "panic":
		tEnd := math.Min(0.55, span*0.9)
		tStep := tEnd / 5.0
		return []map[string]any{
			{"t": t0, "x": baseX, "y": baseY, "w": baseW, "angle": 0.0},
			{"t": math.Round((t0+tStep*1)*1000) / 1000, "x": baseX - 35, "y": baseY - 15, "w": baseW, "angle": -0.12},
			{"t": math.Round((t0+tStep*2)*1000) / 1000, "x": baseX + 35, "y": baseY, "w": baseW, "angle": 0.12},
			{"t": math.Round((t0+tStep*3)*1000) / 1000, "x": baseX - 20, "y": baseY - 10, "w": baseW, "angle": -0.08},
			{"t": math.Round((t0+tStep*4)*1000) / 1000, "x": baseX + 20, "y": baseY, "w": baseW, "angle": 0.08},
			{"t": math.Round((t0+tStep*5)*1000) / 1000, "x": baseX, "y": baseY, "w": baseW, "angle": 0.0},
		}

	case "jump":
		t1 := math.Min(0.22, span*0.45)
		t2 := math.Min(0.38, span*0.75)
		return []map[string]any{
			{"t": t0, "x": baseX, "y": baseY + 180, "w": baseW - 20, "alpha": 0.0},
			{"t": math.Round((t0+t1)*1000) / 1000, "x": baseX, "y": baseY - 40, "w": baseW + 15, "alpha": 1.0},
			{"t": math.Round((t0+t2)*1000) / 1000, "x": baseX, "y": baseY, "w": baseW, "alpha": 1.0},
		}

	case "run":
		t1 := math.Min(0.24, span*0.45)
		t2 := math.Min(0.45, span*0.80)
		return []map[string]any{
			{"t": t0, "x": baseX - 220, "y": baseY, "w": baseW, "angle": 0.14, "alpha": 0.0},
			{"t": math.Round((t0+t1)*1000) / 1000, "x": baseX - 70, "y": baseY - 25, "w": baseW, "angle": 0.10, "alpha": 1.0},
			{"t": math.Round((t0+t2)*1000) / 1000, "x": baseX, "y": baseY, "w": baseW, "angle": 0.0, "alpha": 1.0},
		}

	case "lean":
		t1 := math.Min(0.18, span*0.35)
		t2 := math.Min(0.55, span*0.75)
		t3 := math.Min(0.75, span*0.95)
		return []map[string]any{
			{"t": t0, "x": baseX, "y": baseY, "w": baseW, "angle": 0.0},
			{"t": math.Round((t0+t1)*1000) / 1000, "x": baseX + 20, "y": baseY + 15, "w": baseW, "angle": 0.14},
			{"t": math.Round((t0+t2)*1000) / 1000, "x": baseX + 20, "y": baseY + 15, "w": baseW, "angle": 0.14},
			{"t": math.Round((t0+t3)*1000) / 1000, "x": baseX, "y": baseY, "w": baseW, "angle": 0.0},
		}

	case "punch":
		t1 := math.Min(0.15, span*0.30)
		t2 := math.Min(0.35, span*0.65)
		return []map[string]any{
			{"t": t0, "x": baseX, "y": baseY, "w": baseW},
			{"t": math.Round((t0+t1)*1000) / 1000, "x": baseX - 15, "y": baseY - 30, "w": baseW + 70},
			{"t": math.Round((t0+t2)*1000) / 1000, "x": baseX, "y": baseY, "w": baseW},
		}

	case "peek":
		t1 := math.Min(0.22, span*0.45)
		t2 := math.Min(0.42, span*0.80)
		return []map[string]any{
			{"t": t0, "x": baseX + 160, "y": baseY, "w": baseW, "angle": -0.10, "alpha": 0.0},
			{"t": math.Round((t0+t1)*1000) / 1000, "x": baseX + 40, "y": baseY, "w": baseW, "angle": -0.05, "alpha": 1.0},
			{"t": math.Round((t0+t2)*1000) / 1000, "x": baseX, "y": baseY, "w": baseW, "angle": 0.0, "alpha": 1.0},
		}

	default: // "bounce"
		half := math.Min(0.80, span*0.5)
		return []map[string]any{
			{"t": t0, "x": baseX, "y": baseY, "w": baseW},
			{"t": math.Round((t0+half*0.5)*1000) / 1000, "x": baseX, "y": baseY - 35, "w": baseW + 12},
			{"t": math.Round((t0+half)*1000) / 1000, "x": baseX, "y": baseY, "w": baseW},
		}
	}
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
				sfxList = append(sfxList, map[string]any{
					"time": actionTime,
					"kind": "keyboard",
					"gain": 0.18,
				})
				tKey2 := math.Round((actionTime+0.14)*1000) / 1000
				if tKey2 < totalDuration-0.1 {
					sfxList = append(sfxList, map[string]any{
						"time": tKey2,
						"kind": "keyboard",
						"gain": 0.16,
					})
				}
			case "alert":
				sfxList = append(sfxList, map[string]any{
					"time": actionTime,
					"kind": "alert",
					"gain": 0.22,
				})
			case "glitch":
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
	}

	// 4. Climax / Victory resolution in final scene
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
