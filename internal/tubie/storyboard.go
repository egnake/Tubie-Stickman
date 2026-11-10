package tubie

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

// StickmanActor represents a single stickman character in a scene.
type StickmanActor struct {
	Pose   int    `json:"pose"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	W      int    `json:"w"`
	Action string `json:"action"` // "bounce", "shake", "panic", "jump", "run", "lean", "punch", "peek"
}

// SceneProp represents an icon or hardware asset placed in a scene.
type SceneProp struct {
	Path string `json:"path"`
	X    int    `json:"x"`
	Y    int    `json:"y"`
	W    int    `json:"w"`
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

// BuildStoryboard compiles high-level scene definitions into low-level autopilot shots.
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

		// 1. Scene Title Text
		tFade := math.Min(0.2, shotLen*0.4)
		layers = append(layers, map[string]any{
			"type":  "text",
			"text":  titleText,
			"x":     80,
			"y":     280,
			"w":     1120,
			"size":  58,
			"bold":  true,
			"color": titleColor,
			"keys": []map[string]any{
				{"t": 0.0, "alpha": 0.0},
				{"t": math.Round(tFade*1000) / 1000, "alpha": 1.0},
			},
		})

		// 2. Underline Accent
		if sc.Underline || idx%2 == 0 {
			lineW := len(titleText) * 32
			if lineW < 360 {
				lineW = 360
			}
			if lineW > 1000 {
				lineW = 1000
			}
			tUnderline := math.Min(0.45, shotLen*0.8)
			layers = append(layers, map[string]any{
				"type":      "underline",
				"x":         80,
				"y":         360,
				"w":         lineW,
				"color":     titleColor,
				"lineWidth": 7,
				"keys": []map[string]any{
					{"t": 0.05, "progress": 0.0},
					{"t": math.Round(tUnderline*1000) / 1000, "progress": 1.0},
				},
			})
		}

		// 3. Multi-Stickmen Actors (Poses 0, 1, 2 only)
		stickmen := sc.Stickmen
		if len(stickmen) == 0 {
			switch idx {
			case 0:
				// Hook: startled victim leaps in, menacing hacker peeks
				stickmen = []StickmanActor{
					{Pose: 2, X: 180, Y: 780, W: 470, Action: "jump"},
					{Pose: 1, X: 680, Y: 740, W: 470, Action: "peek"},
				}
			case 1:
				// Action/Trap: victim typing, hacker watching
				stickmen = []StickmanActor{
					{Pose: 0, X: 200, Y: 780, W: 470, Action: "lean"},
					{Pose: 1, X: 680, Y: 740, W: 470, Action: "bounce"},
				}
			case 2:
				// Crisis/Breach: victim in full panic running back and forth
				stickmen = []StickmanActor{
					{Pose: 2, X: 430, Y: 780, W: 500, Action: "panic"},
				}
			case 3:
				// Resolution/Defense: hero defender delivers impact
				stickmen = []StickmanActor{
					{Pose: 1, X: 430, Y: 760, W: 490, Action: "punch"},
				}
			default:
				// Celebration/CTA: lively duo
				stickmen = []StickmanActor{
					{Pose: 1, X: 180, Y: 780, W: 470, Action: "bounce"},
					{Pose: 0, X: 660, Y: 780, W: 470, Action: "jump"},
				}
			}
		}

		for sIdx, sm := range stickmen {
			baseX := sm.X
			if baseX == 0 {
				baseX = 160 + sIdx*500
			}
			baseY := sm.Y
			if baseY == 0 {
				baseY = 780
			}
			baseW := sm.W
			if baseW == 0 {
				baseW = 470
			}

			// Pose must be between 0 and 2
			poseIdx := sm.Pose % 3
			if poseIdx < 0 {
				poseIdx = 0
			}

			keys := buildActorKeyframes(sm.Action, baseX, baseY, baseW, shotLen)

			layers = append(layers, map[string]any{
				"type": "pose",
				"pose": poseIdx,
				"x":    baseX,
				"y":    baseY,
				"w":    baseW,
				"keys": keys,
			})
		}

		// 4. Cyber Prop
		prop := sc.Prop
		if prop == nil {
			prop = &SceneProp{
				Path: availableProps[idx%len(availableProps)],
				X:    490,
				Y:    1380,
				W:    300,
			}
		}
		tProp1 := math.Min(0.25, shotLen*0.4)
		tProp2 := math.Min(0.42, shotLen*0.7)
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

		// 5. Arrow animation
		if sc.Arrow || idx%2 == 1 {
			tArrow := math.Min(0.45, shotLen*0.7)
			layers = append(layers, map[string]any{
				"type":  "arrow",
				"x":     380,
				"y":     1450,
				"w":     520,
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

func buildActorKeyframes(action string, baseX, baseY, baseW int, shotLen float64) []map[string]any {
	switch action {
	case "shake":
		tEnd := math.Min(0.40, shotLen*0.9)
		tStep := tEnd / 5.0
		return []map[string]any{
			{"t": 0.0, "x": baseX, "y": baseY, "w": baseW},
			{"t": math.Round(tStep*1*1000) / 1000, "x": baseX - 14, "y": baseY, "w": baseW},
			{"t": math.Round(tStep*2*1000) / 1000, "x": baseX + 14, "y": baseY, "w": baseW},
			{"t": math.Round(tStep*3*1000) / 1000, "x": baseX - 8, "y": baseY, "w": baseW},
			{"t": math.Round(tStep*4*1000) / 1000, "x": baseX + 8, "y": baseY, "w": baseW},
			{"t": math.Round(tStep*5*1000) / 1000, "x": baseX, "y": baseY, "w": baseW},
		}

	case "panic":
		tEnd := math.Min(0.60, shotLen*0.9)
		tStep := tEnd / 5.0
		return []map[string]any{
			{"t": 0.0, "x": baseX, "y": baseY, "w": baseW, "angle": 0.0},
			{"t": math.Round(tStep*1*1000) / 1000, "x": baseX - 35, "y": baseY - 15, "w": baseW, "angle": -0.12},
			{"t": math.Round(tStep*2*1000) / 1000, "x": baseX + 35, "y": baseY, "w": baseW, "angle": 0.12},
			{"t": math.Round(tStep*3*1000) / 1000, "x": baseX - 20, "y": baseY - 10, "w": baseW, "angle": -0.08},
			{"t": math.Round(tStep*4*1000) / 1000, "x": baseX + 20, "y": baseY, "w": baseW, "angle": 0.08},
			{"t": math.Round(tStep*5*1000) / 1000, "x": baseX, "y": baseY, "w": baseW, "angle": 0.0},
		}

	case "jump":
		t1 := math.Min(0.25, shotLen*0.45)
		t2 := math.Min(0.40, shotLen*0.75)
		return []map[string]any{
			{"t": 0.0, "x": baseX, "y": baseY + 280, "w": baseW - 20, "alpha": 0.0},
			{"t": math.Round(t1*1000) / 1000, "x": baseX, "y": baseY - 40, "w": baseW + 15, "alpha": 1.0},
			{"t": math.Round(t2*1000) / 1000, "x": baseX, "y": baseY, "w": baseW, "alpha": 1.0},
		}

	case "run":
		t1 := math.Min(0.24, shotLen*0.45)
		t2 := math.Min(0.45, shotLen*0.80)
		return []map[string]any{
			{"t": 0.0, "x": baseX - 250, "y": baseY, "w": baseW, "angle": 0.14, "alpha": 0.0},
			{"t": math.Round(t1*1000) / 1000, "x": baseX - 80, "y": baseY - 25, "w": baseW, "angle": 0.10, "alpha": 1.0},
			{"t": math.Round(t2*1000) / 1000, "x": baseX, "y": baseY, "w": baseW, "angle": 0.0, "alpha": 1.0},
		}

	case "lean":
		t1 := math.Min(0.20, shotLen*0.35)
		t2 := math.Min(0.65, shotLen*0.75)
		t3 := math.Min(0.85, shotLen*0.95)
		return []map[string]any{
			{"t": 0.0, "x": baseX, "y": baseY, "w": baseW, "angle": 0.0},
			{"t": math.Round(t1*1000) / 1000, "x": baseX + 20, "y": baseY + 15, "w": baseW, "angle": 0.14},
			{"t": math.Round(t2*1000) / 1000, "x": baseX + 20, "y": baseY + 15, "w": baseW, "angle": 0.14},
			{"t": math.Round(t3*1000) / 1000, "x": baseX, "y": baseY, "w": baseW, "angle": 0.0},
		}

	case "punch":
		t1 := math.Min(0.15, shotLen*0.30)
		t2 := math.Min(0.38, shotLen*0.65)
		return []map[string]any{
			{"t": 0.0, "x": baseX, "y": baseY, "w": baseW},
			{"t": math.Round(t1*1000) / 1000, "x": baseX - 15, "y": baseY - 30, "w": baseW + 70},
			{"t": math.Round(t2*1000) / 1000, "x": baseX, "y": baseY, "w": baseW},
		}

	case "peek":
		t1 := math.Min(0.25, shotLen*0.45)
		t2 := math.Min(0.48, shotLen*0.80)
		return []map[string]any{
			{"t": 0.0, "x": baseX + 180, "y": baseY, "w": baseW, "angle": -0.10, "alpha": 0.0},
			{"t": math.Round(t1*1000) / 1000, "x": baseX + 50, "y": baseY, "w": baseW, "angle": -0.05, "alpha": 1.0},
			{"t": math.Round(t2*1000) / 1000, "x": baseX, "y": baseY, "w": baseW, "angle": 0.0, "alpha": 1.0},
		}

	default: // "bounce"
		half := math.Min(1.0, shotLen*0.5)
		return []map[string]any{
			{"t": 0.0, "x": baseX, "y": baseY, "w": baseW},
			{"t": math.Round(half*0.5*1000) / 1000, "x": baseX, "y": baseY - 35, "w": baseW + 12},
			{"t": math.Round(half*1000) / 1000, "x": baseX, "y": baseY, "w": baseW},
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

		// 3. Action SFX
		actionTime := math.Round((start+math.Min(0.35, shotLen*0.5))*1000) / 1000
		if actionTime < totalDuration-0.1 {
			switch sc.SFX {
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
