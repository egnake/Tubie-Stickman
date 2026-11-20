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
	Title      string          `json:"title"`
	Color      string          `json:"color"`
	Underline  bool            `json:"underline"`
	SFX        string          `json:"sfx"`
	Stickmen   []StickmanActor `json:"stickmen,omitempty"`
	Prop       *SceneProp      `json:"prop,omitempty"`
	Arrow      bool            `json:"arrow"`
	VisualType string          `json:"visual_type,omitempty"`
	Mood       string          `json:"mood,omitempty"`
	Camera     string          `json:"camera,omitempty"`
	Callout    string          `json:"callout,omitempty"`
	VisualData map[string]any  `json:"visual_data,omitempty"`
}

// StoryboardOptions configures aspect-aware layout and acoustic scene synchronization.
type StoryboardOptions struct {
	Width      int
	Height     int
	TimedWords []TimedWord
	Narration  string
}

// ComputeSceneBoundaries aligns scene cut timestamps with real sentence/word boundaries when available.
func ComputeSceneBoundaries(scenes []SceneDefinition, timedWords []TimedWord, narration string, totalDuration float64) []float64 {
	numScenes := len(scenes)
	bounds := make([]float64, numScenes+1)
	if numScenes == 0 {
		return bounds
	}
	bounds[0] = 0.0
	bounds[numScenes] = totalDuration
	if numScenes == 1 || totalDuration <= 0.5 {
		return bounds
	}

	minSpan := math.Min(0.45, (totalDuration/float64(numScenes))*0.60)
	if minSpan <= 0.05 {
		minSpan = 0.05
	}

	usedAcoustic := false
	if len(timedWords) >= numScenes*2 {
		sents := splitSentences(narration)
		wordBoundaries := make([]int, 0, len(sents)+1)
		if len(sents) >= numScenes {
			cumWords := 0
			sentEnds := make([]int, len(sents))
			for i, s := range sents {
				cumWords += len(ExtractScriptWords(s))
				sentEnds[i] = cumWords
			}
			totalScriptWords := sentEnds[len(sentEnds)-1]
			if totalScriptWords > 0 {
				for k := 1; k < numScenes; k++ {
					targetSentIdx := int(math.Round(float64(k)*float64(len(sents))/float64(numScenes))) - 1
					if targetSentIdx < 0 {
						targetSentIdx = 0
					}
					if targetSentIdx >= len(sentEnds) {
						targetSentIdx = len(sentEnds) - 1
					}
					wIdx := int(math.Round(float64(sentEnds[targetSentIdx]) * float64(len(timedWords)) / float64(totalScriptWords)))
					if wIdx < 1 {
						wIdx = 1
					}
					if wIdx >= len(timedWords) {
						wIdx = len(timedWords) - 1
					}
					wordBoundaries = append(wordBoundaries, wIdx)
				}
			}
		} else {
			for k := 1; k < numScenes; k++ {
				wIdx := int(math.Round(float64(k) * float64(len(timedWords)) / float64(numScenes)))
				if wIdx < 1 {
					wIdx = 1
				}
				if wIdx >= len(timedWords) {
					wIdx = len(timedWords) - 1
				}
				wordBoundaries = append(wordBoundaries, wIdx)
			}
		}

		if len(wordBoundaries) == numScenes-1 {
			for k := 1; k < numScenes; k++ {
				wIdx := wordBoundaries[k-1]
				cut := (timedWords[wIdx-1].End + timedWords[wIdx].Start) / 2.0
				minAllowed := bounds[k-1] + minSpan
				maxAllowed := totalDuration - float64(numScenes-k)*minSpan
				if cut < minAllowed {
					cut = minAllowed
				}
				if cut > maxAllowed {
					cut = maxAllowed
				}
				bounds[k] = math.Round(cut*1000) / 1000
			}
			usedAcoustic = true
		}
	}

	if !usedAcoustic {
		sceneDur := totalDuration / float64(numScenes)
		for k := 1; k < numScenes; k++ {
			t := math.Round(float64(k)*sceneDur*1000) / 1000
			minAllowed := bounds[k-1] + minSpan
			maxAllowed := totalDuration - float64(numScenes-k)*minSpan
			if t < minAllowed {
				t = minAllowed
			}
			if t > maxAllowed {
				t = maxAllowed
			}
			bounds[k] = t
		}
	}

	return bounds
}

// BuildStoryboard compiles high-level scene definitions into rich autopilot shots with micro-acting and cyber staging.
func BuildStoryboard(scenes []SceneDefinition, totalDuration float64) []map[string]any {
	return BuildStoryboardWithOptions(scenes, totalDuration, StoryboardOptions{Width: 720, Height: 1280})
}

// BuildStoryboardWithOptions compiles scenes into aspect-aware, acoustically synced procedural cyber shots.
func BuildStoryboardWithOptions(scenes []SceneDefinition, totalDuration float64, opts StoryboardOptions) []map[string]any {
	if len(scenes) == 0 {
		return []map[string]any{}
	}
	if opts.Width <= 0 || opts.Height <= 0 {
		opts.Width, opts.Height = 720, 1280
	}
	isLandscape := opts.Width > opts.Height

	numScenes := len(scenes)
	bounds := ComputeSceneBoundaries(scenes, opts.TimedWords, opts.Narration, totalDuration)
	shots := make([]map[string]any, 0, numScenes)

	for idx, rawSc := range scenes {
		sc := inferSceneVisuals(rawSc, idx, numScenes, "", opts.Narration)
		start := bounds[idx]
		end := bounds[idx+1]
		if end <= start {
			end = start + 0.5
		}
		rawSpan := end - start
		shotLen := math.Floor(rawSpan*1000) / 1000
		if shotLen <= 0.05 {
			shotLen = rawSpan
		}
		tSwitch := math.Floor(math.Min(shotLen*0.44, 1.85)*1000) / 1000
		if tSwitch >= shotLen-0.08 {
			tSwitch = math.Floor(shotLen*0.45*1000) / 1000
		}

		titleText := strings.ToUpper(strings.TrimSpace(sc.Title))
		if titleText == "" {
			titleText = fmt.Sprintf("BÖLÜM 0%d", idx+1)
		}

		titleColor := sc.Color
		if titleColor == "" {
			colors := []string{"#00F0FF", "#FF3B30", "#FFCC00", "#00FF66", "#E056FD"}
			titleColor = colors[idx%len(colors)]
		}

		layers := make([]map[string]any, 0, 10)

		// Coordinates adapt cleanly to 9:16 Portrait (1280x2275) vs 16:9 Landscape (1280x720)
		hudY, titleX, titleY, titleW, titleSize := 112, 72, 165, 1136, 56
		underlineY, pillX, pillY, pillSize := 288, 640, 336, 22
		if isLandscape {
			hudY, titleX, titleY, titleW, titleSize = 72, 54, 102, 1172, 42
			underlineY, pillX, pillY, pillSize = 154, 960, 72, 18
		}

		// 1. Cyber Security HUD Header Badge (immediately visible at t=0.0 — zero blank frames)
		hudText, hudColor := getHUDBadge(idx, titleText)
		layers = append(layers, map[string]any{
			"type":  "text",
			"text":  hudText,
			"x":     titleX,
			"y":     hudY,
			"w":     titleW,
			"size":  26,
			"bold":  true,
			"color": hudColor,
			"keys": buildSnappyAppearKeys(titleX, hudY, 0, shotLen),
		})

		// 2. Main Scene Title Text (instant punch-in, never starts invisible)
		layers = append(layers, map[string]any{
			"type":  "text",
			"text":  titleText,
			"x":     titleX,
			"y":     titleY,
			"w":     titleW,
			"size":  titleSize,
			"bold":  true,
			"color": titleColor,
			"keys":  buildSnappySlideKeys(titleX, titleY, titleY+8, shotLen),
		})

		// 3. Neon Underline Accent
		if sc.Underline || idx%2 == 0 {
			lineW := len([]rune(titleText)) * 31
			if isLandscape {
				lineW = len([]rune(titleText)) * 22
			}
			if lineW < 340 {
				lineW = 340
			}
			if lineW > 1020 {
				lineW = 1020
			}
			tUnderline := math.Floor(math.Min(0.28, shotLen*0.70)*1000) / 1000
			if tUnderline < 0.02 {
				tUnderline = shotLen * 0.5
			}
			layers = append(layers, map[string]any{
				"type":      "underline",
				"x":         titleX,
				"y":         underlineY,
				"w":         lineW,
				"color":     titleColor,
				"lineWidth": 7,
				"keys": buildProgressKeys(0.0, tUnderline, 0.25, 1.0, shotLen),
			})
		}

		// 4. Dynamic Cyber Status Pill Badge
		pillText, pillColor := getStatusPill(idx, titleText)
		pillLayer := map[string]any{
			"type":  "text",
			"text":  pillText,
			"x":     pillX,
			"y":     pillY,
			"w":     860,
			"size":  pillSize,
			"bold":  true,
			"color": pillColor,
			"badge": pillColor,
			"keys":  buildSnappySlideKeys(pillX, pillY, pillY+6, shotLen),
		}
		if !isLandscape {
			pillLayer["center"] = true
		}
		layers = append(layers, pillLayer)

		// 5. Procedural Cyber Stickman & Hooded Hacker Actors
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
			layers = append(layers, buildProceduralCyberActors(sc, idx, shotLen, tSwitch, isLandscape)...)
		}

		// 6. Live Animated Cyber Widget (or staged prop)
		layers = append(layers, buildCyberWidgetLayer(sc, shotLen, isLandscape))

		// 7. Pattern-Interrupt Callout Stamp (pops cleanly at tSwitch on shots long enough)
		if sc.Callout != "" && shotLen >= 0.85 {
			cStart := tSwitch
			if cStart <= 0.05 {
				cStart = math.Floor(shotLen*0.35*1000) / 1000
			}
			if cStart < shotLen-0.12 {
				cx, cy, cw, csize := 440, 425, 400, 36
				if isLandscape {
					cx, cy, cw, csize = 80, 178, 380, 26
				}
				layers = append(layers, map[string]any{
					"type":  "cyber_callout",
					"text":  sc.Callout,
					"x":     cx,
					"y":     cy,
					"w":     cw,
					"size":  csize,
					"color": titleColor,
					"from":  cStart,
				})
			}
		}

		// 8. Optional Threat Direction Arrow on danger/breach scenes
		if sc.Arrow || sc.Mood == "danger" || sc.Mood == "breach" {
			ax, ay, aw := 340, 480, 600
			if isLandscape {
				ax, ay, aw = 180, 215, 220
			}
			tArrow := math.Floor(math.Min(0.38, shotLen*0.65)*1000) / 1000
			layers = append(layers, map[string]any{
				"type":      "arrow",
				"x":         ax,
				"y":         ay,
				"w":         aw,
				"color":     "#FF3B4E",
				"lineWidth": 5,
				"keys":      buildProgressKeys(0.0, tArrow, 0.0, 1.0, shotLen),
			})
		}

		shots = append(shots, map[string]any{
			"start":  start,
			"end":    end,
			"mood":   sc.Mood,
			"camera": sc.Camera,
			"layers": layers,
		})
	}

	return shots
}

func buildSnappyAppearKeys(x, y int, startT, shotLen float64) []map[string]any {
	rt0 := math.Floor(startT*1000) / 1000
	rt1 := math.Floor(math.Min(startT+0.12, shotLen*0.5)*1000) / 1000
	keys := []map[string]any{
		{"t": rt0, "x": x, "y": y, "alpha": 0.85},
	}
	if rt1 > rt0 && rt1 <= shotLen {
		keys = append(keys, map[string]any{"t": rt1, "x": x, "y": y, "alpha": 1.0})
	}
	return keys
}

func buildSnappySlideKeys(x, targetY, startY int, shotLen float64) []map[string]any {
	rt1 := math.Floor(math.Min(0.14, shotLen*0.45)*1000) / 1000
	keys := []map[string]any{
		{"t": 0.0, "x": x, "y": startY, "alpha": 0.85},
	}
	if rt1 > 0.0 && rt1 <= shotLen {
		keys = append(keys, map[string]any{"t": rt1, "x": x, "y": targetY, "alpha": 1.0})
	}
	return keys
}

func buildProgressKeys(t0, t1, p0, p1, shotLen float64) []map[string]any {
	rt0 := math.Floor(t0*1000) / 1000
	rt1 := math.Floor(t1*1000) / 1000
	if rt1 > shotLen {
		rt1 = math.Floor(shotLen*1000) / 1000
	}
	keys := []map[string]any{{"t": rt0, "progress": p0}}
	if rt1 > rt0 {
		keys = append(keys, map[string]any{"t": rt1, "progress": p1})
	}
	return keys
}

func buildProceduralCyberActors(sc SceneDefinition, idx int, shotLen, tSwitch float64, isLandscape bool) []map[string]any {
	// Portrait coordinates (1280x2275 plane) vs Landscape coordinates (1280x720 plane)
	vx, vy, vw, vh := 130, 445, 420, 510
	hx, hy, hw, hh := 710, 425, 440, 530
	cx, cy, cw, ch := 415, 435, 450, 520
	if isLandscape {
		vx, vy, vw, vh = 40, 195, 250, 355
		hx, hy, hw, hh = 275, 185, 265, 365
		cx, cy, cw, ch = 145, 185, 275, 365
	}

	canSplit := shotLen >= 0.75 && tSwitch > 0.12 && tSwitch < shotLen-0.12
	layers := make([]map[string]any, 0, 4)

	switch sc.Mood {
	case "hook":
		if canSplit {
			layers = append(layers,
				map[string]any{
					"type": "cyber_actor", "role": "victim", "pose": "idle_browse",
					"x": vx, "y": vy, "w": vw, "h": vh, "to": tSwitch,
					"keys": buildMotionKeys(vx, vy, vw, 0, tSwitch, "lean"),
				},
				map[string]any{
					"type": "cyber_actor", "role": "victim", "pose": "shocked",
					"x": vx, "y": vy, "w": vw, "h": vh, "from": tSwitch,
					"keys": buildMotionKeys(vx, vy, vw, tSwitch, shotLen-tSwitch, "jump"),
				},
				map[string]any{
					"type": "cyber_actor", "role": "hacker", "pose": "hacker_lurk",
					"x": hx, "y": hy, "w": hw, "h": hh, "from": tSwitch,
					"keys": buildMotionKeys(hx, hy, hw, tSwitch, shotLen-tSwitch, "peek"),
				},
			)
		} else {
			layers = append(layers, map[string]any{
				"type": "cyber_actor", "role": "victim", "pose": "shocked",
				"x": cx, "y": cy, "w": cw, "h": ch,
				"keys": buildMotionKeys(cx, cy, cw, 0, shotLen, "bounce"),
			})
		}

	case "danger":
		if canSplit {
			layers = append(layers,
				map[string]any{
					"type": "cyber_actor", "role": "victim", "pose": "typing",
					"x": vx, "y": vy, "w": vw, "h": vh, "to": tSwitch,
					"keys": buildMotionKeys(vx, vy, vw, 0, tSwitch, "lean"),
				},
				map[string]any{
					"type": "cyber_actor", "role": "victim", "pose": "shocked",
					"x": vx, "y": vy, "w": vw, "h": vh, "from": tSwitch,
					"keys": buildMotionKeys(vx, vy, vw, tSwitch, shotLen-tSwitch, "panic"),
				},
				map[string]any{
					"type": "cyber_actor", "role": "hacker", "pose": "hacker_strike",
					"x": hx, "y": hy, "w": hw, "h": hh,
					"keys": buildMotionKeys(hx, hy, hw, 0, shotLen, "punch"),
				},
			)
		} else {
			layers = append(layers,
				map[string]any{
					"type": "cyber_actor", "role": "victim", "pose": "shocked",
					"x": vx, "y": vy, "w": vw, "h": vh,
					"keys": buildMotionKeys(vx, vy, vw, 0, shotLen, "panic"),
				},
				map[string]any{
					"type": "cyber_actor", "role": "hacker", "pose": "hacker_strike",
					"x": hx, "y": hy, "w": hw, "h": hh,
					"keys": buildMotionKeys(hx, hy, hw, 0, shotLen, "punch"),
				},
			)
		}

	case "breach":
		layers = append(layers,
			map[string]any{
				"type": "cyber_actor", "role": "victim", "pose": "panic_run",
				"x": vx, "y": vy, "w": vw, "h": vh,
				"keys": buildMotionKeys(vx, vy, vw, 0, shotLen, "run"),
			},
			map[string]any{
				"type": "cyber_actor", "role": "hacker", "pose": "hacker_strike",
				"x": hx, "y": hy, "w": hw, "h": hh,
				"keys": buildMotionKeys(hx, hy, hw, 0, shotLen, "punch"),
			},
		)

	case "twist":
		if canSplit {
			layers = append(layers,
				map[string]any{
					"type": "cyber_actor", "role": "victim", "pose": "thinking",
					"x": cx, "y": cy, "w": cw, "h": ch, "to": tSwitch,
					"keys": buildMotionKeys(cx, cy, cw, 0, tSwitch, "shake"),
				},
				map[string]any{
					"type": "cyber_actor", "role": "hero", "pose": "pointing",
					"x": cx, "y": cy, "w": cw, "h": ch, "from": tSwitch,
					"keys": buildMotionKeys(cx, cy, cw, tSwitch, shotLen-tSwitch, "bounce"),
				},
			)
		} else {
			layers = append(layers, map[string]any{
				"type": "cyber_actor", "role": "victim", "pose": "thinking",
				"x": cx, "y": cy, "w": cw, "h": ch,
				"keys": buildMotionKeys(cx, cy, cw, 0, shotLen, "shake"),
			})
		}

	case "defense":
		if canSplit {
			layers = append(layers,
				map[string]any{
					"type": "cyber_actor", "role": "hero", "pose": "pointing",
					"x": cx, "y": cy, "w": cw, "h": ch, "to": tSwitch,
					"keys": buildMotionKeys(cx, cy, cw, 0, tSwitch, "lean"),
				},
				map[string]any{
					"type": "cyber_actor", "role": "hero", "pose": "shield_hero",
					"x": cx, "y": cy, "w": cw, "h": ch, "from": tSwitch,
					"keys": buildMotionKeys(cx, cy, cw, tSwitch, shotLen-tSwitch, "jump"),
				},
			)
		} else {
			layers = append(layers, map[string]any{
				"type": "cyber_actor", "role": "hero", "pose": "shield_hero",
				"x": cx, "y": cy, "w": cw, "h": ch,
				"keys": buildMotionKeys(cx, cy, cw, 0, shotLen, "bounce"),
			})
		}

	default: // "cta"
		layers = append(layers,
			map[string]any{
				"type": "cyber_actor", "role": "hero", "pose": "shield_hero",
				"x": vx, "y": vy, "w": vw, "h": vh,
				"keys": buildMotionKeys(vx, vy, vw, 0, shotLen, "bounce"),
			},
			map[string]any{
				"type": "cyber_actor", "role": "hero", "pose": "celebrate",
				"x": hx, "y": hy, "w": hw, "h": hh,
				"keys": buildMotionKeys(hx, hy, hw, 0, shotLen, "jump"),
			},
		)
	}
	return layers
}

func buildMotionKeys(baseX, baseY, baseW int, tStart, span float64, action string) []map[string]any {
	keys := buildActorKeyframes(action, baseX, baseY, baseW, tStart, span)
	// Never begin a primary actor at alpha=0 on opening frame t=0.0; clamp alpha >= 0.9 on t==0
	if len(keys) > 0 {
		if tVal, ok := keys[0]["t"].(float64); ok && tVal == 0.0 {
			if a, hasA := keys[0]["alpha"].(float64); hasA && a < 0.85 {
				keys[0]["alpha"] = 0.90
			}
		}
	}
	return keys
}

func buildCyberWidgetLayer(sc SceneDefinition, shotLen float64, isLandscape bool) map[string]any {
	wx, wy, ww, wh := 70, 1020, 1140, 520
	if isLandscape {
		wx, wy, ww, wh = 565, 175, 670, 375
	}

	if sc.VisualType == "prop_stage" && sc.Prop != nil && sc.Prop.Path != "" {
		pw := 460
		px := 410
		py := 1050
		if isLandscape {
			pw = 340
			px = 730
			py = 195
		}
		return map[string]any{
			"type": "image",
			"path": sc.Prop.Path,
			"x":    px,
			"y":    py,
			"w":    pw,
			"glow": "cyan",
			"keys": buildSnappySlideKeys(px, py, py+16, shotLen),
		}
	}

	kind := sc.VisualType
	if !validVisualTypes[kind] || kind == "prop_stage" {
		kind = "terminal"
	}
	return map[string]any{
		"type": "cyber_widget",
		"kind": kind,
		"x":    wx,
		"y":    wy,
		"w":    ww,
		"h":    wh,
		"data": sc.VisualData,
		"keys": buildSnappySlideKeys(wx, wy, wy+18, shotLen),
	}
}

func getHUDBadge(idx int, title string) (string, string) {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "yapay zeka") || strings.Contains(lower, "gizli") || strings.Contains(lower, "sohbet") || strings.Contains(lower, "profil") || strings.Contains(lower, "dijital iz"):
		return "[ DİJİTAL GİZLİLİK // VERİ ANALİZİ ]", "#00F0FF"
	case strings.Contains(lower, "kripto") || strings.Contains(lower, "cüzdan") || strings.Contains(lower, "airdrop") || strings.Contains(lower, "token"):
		return "[ KRİPTO GÜVENLİK // CÜZDAN ALARMI ]", "#FFCC00"
	case strings.Contains(lower, "tuzak") || strings.Contains(lower, "sahte") || strings.Contains(lower, "dm") || strings.Contains(lower, "oltalama"):
		return "[ ALERT // OLTALAMA SALDIRISI ]", "#FF3B30"
	case strings.Contains(lower, "ele geçiril") || strings.Contains(lower, "çalındı") || strings.Contains(lower, "şifre") || strings.Contains(lower, "parola"):
		return "[ SİSTEM İHLALİ // KRİTİK UYARI ]", "#FF3B30"
	case strings.Contains(lower, "kaybedil") || strings.Contains(lower, "kontrol") || strings.Contains(lower, "sızma") || strings.Contains(lower, "takip"):
		return "[ KRİTİK ALARM // RİSK TESPİTİ ]", "#FFCC00"
	case strings.Contains(lower, "istemez") || strings.Contains(lower, "gerçek") || strings.Contains(lower, "farkındalık") || strings.Contains(lower, "halka"):
		return "[ SİBER GERÇEK // TUZAĞA DÜŞME ]", "#E056FD"
	case strings.Contains(lower, "koru") || strings.Contains(lower, "2fa") || strings.Contains(lower, "doğrulama") || strings.Contains(lower, "kalkan") || strings.Contains(lower, "kapat"):
		return "[ GÜVENLİK DUVARI // AKTİF SAVUNMA ]", "#00FF66"
	case strings.Contains(lower, "uyanık") || strings.Contains(lower, "güvende") || strings.Contains(lower, "avcı"):
		return "[ KORUMA AKTİF // GÜVENDESİN ]", "#00F0FF"
	default:
		hudList := []struct {
			text  string
			color string
		}{
			{"[ SİBER İSTİHBARAT // VERİ ANALİZİ ]", "#00F0FF"},
			{"[ TEHLİKE DÜZEYİ // KRİTİK ALARM ]", "#FF3B30"},
			{"[ SİSTEM ANALİZİ // RİSK TESPİTİ ]", "#FFCC00"},
			{"[ GÜVENLİK GERÇEĞİ // DİKKAT ]", "#E056FD"},
			{"[ AKTİF SAVUNMA // SİBER KALKAN ]", "#00FF66"},
			{"[ GÜVENDE KAL // BİLİNÇLİ KULLANICI ]", "#00F0FF"},
		}
		item := hudList[idx%len(hudList)]
		return item.text, item.color
	}
}

func getStatusPill(idx int, title string) (string, string) {
	lower := strings.ToLower(title)
	switch {
	case strings.Contains(lower, "yapay zeka") || strings.Contains(lower, "gizli") || strings.Contains(lower, "sohbet") || strings.Contains(lower, "profil") || strings.Contains(lower, "dijital iz"):
		return "[ DİJİTAL İZ VE VERİ TOPLAMA ANALİZİ ]", "#00F0FF"
	case strings.Contains(lower, "kripto") || strings.Contains(lower, "cüzdan") || strings.Contains(lower, "airdrop") || strings.Contains(lower, "token"):
		return "[ UYARI: AKILLI SÖZLEŞME VE CÜZDAN RİSKİ ]", "#FFCC00"
	case strings.Contains(lower, "tuzak") || strings.Contains(lower, "sahte") || strings.Contains(lower, "dm") || strings.Contains(lower, "oltalama"):
		return "[ TEHLİKELİ BAĞLANTI TESPİT EDİLDİ ]", "#FF3B30"
	case strings.Contains(lower, "ele geçiril") || strings.Contains(lower, "çalındı") || strings.Contains(lower, "şifre") || strings.Contains(lower, "parola"):
		return "[ ALARM: KİMLİK BİLGİLERİ RİSK ALTINDA ]", "#FF3B30"
	case strings.Contains(lower, "kaybedil") || strings.Contains(lower, "kontrol") || strings.Contains(lower, "sızma") || strings.Contains(lower, "takip"):
		return "[ KRİTİK İHLAL: DİJİTAL KONTROL RİSKİ ]", "#FFCC00"
	case strings.Contains(lower, "istemez") || strings.Contains(lower, "gerçek") || strings.Contains(lower, "farkındalık") || strings.Contains(lower, "halka"):
		return "[ SİBER GERÇEK: EN ZAYIF HALKA İNSAN ]", "#E056FD"
	case strings.Contains(lower, "koru") || strings.Contains(lower, "2fa") || strings.Contains(lower, "doğrulama") || strings.Contains(lower, "kalkan") || strings.Contains(lower, "kapat"):
		return "[ GÜVENLİK KALKANI: ÖNLEMLER DEVREDE ]", "#00FF66"
	case strings.Contains(lower, "uyanık") || strings.Contains(lower, "güvende") || strings.Contains(lower, "avcı"):
		return "[ TAM SAVUNMA: BİLİNÇLİ KULLANICI ]", "#00F0FF"
	default:
		pills := []struct {
			text  string
			color string
		}{
			{"[ SİBER GÜVENLİK ANALİZİ BAŞLATILDI ]", "#00F0FF"},
			{"[ KRİTİK DİJİTAL RİSK UYARISI ]", "#FF3B30"},
			{"[ ALARM: GİZLİ VERİ İHLALİ RİSKİ ]", "#FFCC00"},
			{"[ DİKKAT: VERİ PAYLAŞIMINI KONTROL ET ]", "#E056FD"},
			{"[ GÜVENLİK KALKANI DEVREDE ]", "#00FF66"},
			{"[ BİLİNÇLİ KULLANICI // TAM KORUMA ]", "#00F0FF"},
		}
		item := pills[idx%len(pills)]
		return item.text, item.color
	}
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
		addKey(t0, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": 0.0})
		addKey(t0+0.08, map[string]any{"x": baseX - 24, "y": baseY - 16, "w": baseW + 12, "angle": -0.10})
		addKey(t0+0.16, map[string]any{"x": baseX + 24, "y": baseY - 4, "w": baseW + 8, "angle": 0.10})
		addKey(t0+0.24, map[string]any{"x": baseX - 14, "y": baseY - 10, "w": baseW, "angle": -0.06})
		addKey(t0+0.32, map[string]any{"x": baseX + 14, "y": baseY, "w": baseW, "angle": 0.06})
		addKey(t0+0.40, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": 0.0})

		step := 0.16
		cur := t0 + 0.40 + step
		toggle := false
		for cur < maxT-0.05 {
			jitterX := baseX + 7
			jitterAngle := 0.03
			if toggle {
				jitterX = baseX - 7
				jitterAngle = -0.03
			}
			addKey(cur, map[string]any{
				"x":     jitterX,
				"y":     baseY + int(math.Sin(cur*15.0)*4.0),
				"w":     baseW,
				"angle": jitterAngle,
			})
			toggle = !toggle
			cur += step
		}
		addKey(maxT, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": 0.0})

	case "shake":
		addKey(t0, map[string]any{"x": baseX, "y": baseY, "w": baseW})
		addKey(t0+0.08, map[string]any{"x": baseX - 14, "y": baseY, "w": baseW})
		addKey(t0+0.16, map[string]any{"x": baseX + 14, "y": baseY, "w": baseW})
		addKey(t0+0.24, map[string]any{"x": baseX - 9, "y": baseY, "w": baseW})
		addKey(t0+0.32, map[string]any{"x": baseX + 9, "y": baseY, "w": baseW})
		addKey(t0+0.40, map[string]any{"x": baseX, "y": baseY, "w": baseW})

		step := 0.42
		cur := t0 + 0.40 + step
		for cur < maxT-0.05 {
			addKey(cur, map[string]any{"x": baseX, "y": baseY - 10, "w": baseW + 5})
			if cur+step*0.5 < maxT-0.05 {
				addKey(cur+step*0.5, map[string]any{"x": baseX, "y": baseY, "w": baseW})
			}
			cur += step
		}
		addKey(maxT, map[string]any{"x": baseX, "y": baseY, "w": baseW})

	case "jump":
		t1 := math.Min(0.18, maxSpan*0.40)
		t2 := math.Min(0.34, maxSpan*0.70)
		addKey(t0, map[string]any{"x": baseX, "y": baseY + 60, "w": baseW - 14, "alpha": 0.9})
		addKey(t0+t1, map[string]any{"x": baseX, "y": baseY - 32, "w": baseW + 16, "alpha": 1.0})
		addKey(t0+t2, map[string]any{"x": baseX, "y": baseY, "w": baseW, "alpha": 1.0})

		step := 0.42
		cur := t0 + t2 + step
		for cur < maxT-0.05 {
			addKey(cur, map[string]any{"x": baseX, "y": baseY - 14, "w": baseW + 6})
			if cur+step*0.5 < maxT-0.05 {
				addKey(cur+step*0.5, map[string]any{"x": baseX, "y": baseY, "w": baseW})
			}
			cur += step
		}
		addKey(maxT, map[string]any{"x": baseX, "y": baseY, "w": baseW, "alpha": 1.0})

	case "run":
		t1 := math.Min(0.20, maxSpan*0.40)
		t2 := math.Min(0.40, maxSpan*0.75)
		addKey(t0, map[string]any{"x": baseX - 90, "y": baseY, "w": baseW, "angle": 0.10, "alpha": 0.9})
		addKey(t0+t1, map[string]any{"x": baseX - 30, "y": baseY - 20, "w": baseW, "angle": 0.06, "alpha": 1.0})
		addKey(t0+t2, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": 0.0, "alpha": 1.0})

		step := 0.35
		cur := t0 + t2 + step
		toggle := false
		for cur < maxT-0.05 {
			swayX := baseX + 12
			if toggle {
				swayX = baseX - 12
			}
			addKey(cur, map[string]any{
				"x":     swayX,
				"y":     baseY - 12,
				"w":     baseW,
				"angle": 0.03 * (map[bool]float64{true: -1, false: 1}[toggle]),
			})
			toggle = !toggle
			cur += step
		}
		addKey(maxT, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": 0.0, "alpha": 1.0})

	case "lean":
		addKey(t0, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": 0.0})
		addKey(t0+0.14, map[string]any{"x": baseX + 12, "y": baseY + 8, "w": baseW, "angle": 0.07})

		step := 0.42
		cur := t0 + 0.14 + step
		toggle := false
		for cur < maxT-0.05 {
			leanAngle := 0.09
			leanY := baseY + 10
			if toggle {
				leanAngle = 0.05
				leanY = baseY + 6
			}
			addKey(cur, map[string]any{
				"x":     baseX + 12,
				"y":     leanY,
				"w":     baseW,
				"angle": leanAngle,
			})
			toggle = !toggle
			cur += step
		}
		addKey(maxT, map[string]any{"x": baseX + 12, "y": baseY + 8, "w": baseW, "angle": 0.07})

	case "punch":
		addKey(t0, map[string]any{"x": baseX, "y": baseY, "w": baseW})
		addKey(t0+0.12, map[string]any{"x": baseX - 14, "y": baseY - 18, "w": baseW + 28})
		addKey(t0+0.24, map[string]any{"x": baseX, "y": baseY, "w": baseW})

		step := 0.24
		cur := t0 + 0.24 + step
		toggle := false
		for cur < maxT-0.05 {
			strikeY := baseY - 14
			strikeW := baseW + 20
			strikeX := baseX - 8
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
		t1 := math.Min(0.22, maxSpan*0.45)
		t2 := math.Min(0.44, maxSpan*0.80)
		addKey(t0, map[string]any{"x": baseX + 100, "y": baseY, "w": baseW, "angle": -0.07, "alpha": 0.9})
		addKey(t0+t1, map[string]any{"x": baseX + 30, "y": baseY - 10, "w": baseW, "angle": -0.04, "alpha": 1.0})
		addKey(t0+t2, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": -0.02, "alpha": 1.0})

		step := 0.45
		cur := t0 + t2 + step
		toggle := false
		for cur < maxT-0.05 {
			addKey(cur, map[string]any{
				"x":     baseX + map[bool]int{true: 10, false: -6}[toggle],
				"y":     baseY + map[bool]int{true: -8, false: 0}[toggle],
				"w":     baseW,
				"angle": -0.02,
			})
			toggle = !toggle
			cur += step
		}
		addKey(maxT, map[string]any{"x": baseX, "y": baseY, "w": baseW, "angle": -0.02, "alpha": 1.0})

	default: // "bounce"
		step := 0.42
		cur := t0
		toggle := false
		addKey(t0, map[string]any{"x": baseX, "y": baseY, "w": baseW})
		cur += step * 0.5
		for cur < maxT-0.05 {
			bY := baseY - 22
			bW := baseW + 10
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
	return BuildSoundDesignWithTimings(scenes, nil, totalDuration)
}

// BuildSoundDesignWithTimings places synchronized SFX cues aligned to exact scene cut boundaries.
func BuildSoundDesignWithTimings(scenes []SceneDefinition, boundaries []float64, totalDuration float64) []map[string]any {
	if len(scenes) == 0 || totalDuration <= 0.5 {
		return []map[string]any{}
	}

	numScenes := len(scenes)
	if len(boundaries) != numScenes+1 {
		boundaries = ComputeSceneBoundaries(scenes, nil, "", totalDuration)
	}
	sfxList := make([]map[string]any, 0, numScenes*5)

	// Opening hook punch (0.02s)
	sfxList = append(sfxList, map[string]any{
		"time": 0.02,
		"kind": "pop",
		"gain": 0.20,
	})

	for i, rawSc := range scenes {
		sc := inferSceneVisuals(rawSc, i, numScenes, "", "")
		start := boundaries[i]
		end := boundaries[i+1]
		shotLen := end - start
		if shotLen <= 0.05 {
			continue
		}
		tSwitch := math.Round(math.Min(shotLen*0.44, 1.85)*1000) / 1000

		// 1. Scene transition whoosh/pop
		if i > 0 {
			tCut := math.Round((start+0.02)*1000) / 1000
			if tCut < totalDuration-0.1 {
				sfxList = append(sfxList, map[string]any{
					"time": tCut,
					"kind": "paper",
					"gain": 0.14,
				})
			}
		}

		// 2. Pre-twist/danger riser
		if i > 0 && (sc.Mood == "danger" || sc.Mood == "breach" || sc.Mood == "twist") && start > 0.65 {
			tRise := math.Round((start-0.45)*1000) / 1000
			if tRise >= 0.05 && tRise < totalDuration-0.15 {
				sfxList = append(sfxList, map[string]any{
					"time": tRise,
					"kind": "pop",
					"gain": 0.14,
				})
			}
		}

		// 3. Scene Event Reaction SFX
		actionTime := math.Round((start+tSwitch)*1000) / 1000
		if actionTime < totalDuration-0.1 {
			kind := sc.SFX
			if kind == "" {
				switch sc.Mood {
				case "hook":
					kind = "keyboard"
				case "danger":
					kind = "alert"
				case "breach":
					kind = "glitch"
				case "twist":
					kind = "chime"
				case "defense":
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
			case "alert", "glitch", "lock", "chime", "jingle", "paper", "pop", "tap":
				gain := 0.19
				if kind == "alert" || kind == "glitch" || kind == "lock" || kind == "jingle" {
					gain = 0.21
				}
				sfxList = append(sfxList, map[string]any{
					"time": actionTime,
					"kind": kind,
					"gain": gain,
				})
			default:
				sfxList = append(sfxList, map[string]any{
					"time": actionTime,
					"kind": "tap",
					"gain": 0.16,
				})
			}
		}

		// 4. Micro-reaction / callout pop cue
		tSec := math.Round((start+tSwitch+0.16)*1000) / 1000
		if tSec < totalDuration-0.1 {
			sfxList = append(sfxList, map[string]any{
				"time": tSec,
				"kind": "tap",
				"gain": 0.14,
			})
		}
	}

	// 5. Climax / Victory resolution in final scene
	tClimax := math.Round((totalDuration-0.55)*1000) / 1000
	if tClimax > 0.4 {
		sfxList = append(sfxList, map[string]any{
			"time": tClimax,
			"kind": "jingle",
			"gain": 0.20,
		})
	}

	sort.Slice(sfxList, func(i, j int) bool {
		return sfxList[i]["time"].(float64) < sfxList[j]["time"].(float64)
	})

	return sfxList
}
