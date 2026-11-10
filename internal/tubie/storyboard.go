package tubie

import (
	"fmt"
	"math"
	"strings"
)

// StickmanActor represents a single stickman character in a scene.
type StickmanActor struct {
	Pose   int    `json:"pose"`
	X      int    `json:"x"`
	Y      int    `json:"y"`
	W      int    `json:"w"`
	Action string `json:"action"` // "bounce" or "shake"
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
			colors := []string{"#00F0FF", "#FF3B30", "#FFCC00", "#00FF66"}
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
			// Auto assign 2 interactive stickmen
			stickmen = []StickmanActor{
				{Pose: idx % 3, X: 140, Y: 800, W: 470, Action: "bounce"},
				{Pose: (idx + 1) % 3, X: 670, Y: 790, W: 470, Action: "shake"},
			}
		}

		for sIdx, sm := range stickmen {
			baseX := sm.X
			if baseX == 0 {
				baseX = 160 + sIdx*500
			}
			baseY := sm.Y
			if baseY == 0 {
				baseY = 800
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

			var keys []map[string]any
			if sm.Action == "shake" {
				tEndShake := math.Min(0.40, shotLen*0.9)
				tStep := tEndShake / 5.0
				keys = []map[string]any{
					{"t": 0.0, "x": baseX, "y": baseY, "w": baseW},
					{"t": math.Round(tStep*1*1000) / 1000, "x": baseX - 14, "y": baseY, "w": baseW},
					{"t": math.Round(tStep*2*1000) / 1000, "x": baseX + 14, "y": baseY, "w": baseW},
					{"t": math.Round(tStep*3*1000) / 1000, "x": baseX - 8, "y": baseY, "w": baseW},
					{"t": math.Round(tStep*4*1000) / 1000, "x": baseX + 8, "y": baseY, "w": baseW},
					{"t": math.Round(tStep*5*1000) / 1000, "x": baseX, "y": baseY, "w": baseW},
				}
			} else {
				half := math.Min(1.0, shotLen*0.5)
				keys = []map[string]any{
					{"t": 0.0, "x": baseX, "y": baseY, "w": baseW},
					{"t": math.Round(half*0.5*1000) / 1000, "x": baseX, "y": baseY - 30, "w": baseW + 12},
					{"t": math.Round(half*1000) / 1000, "x": baseX, "y": baseY, "w": baseW},
				}
			}

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
				Y:    1420,
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
				{"t": 0.0, "w": 80, "alpha": 0.0, "angle": 0.0},
				{"t": math.Round(tProp1*1000) / 1000, "w": prop.W + 35, "alpha": 1.0, "angle": 0.05},
				{"t": math.Round(tProp2*1000) / 1000, "w": prop.W, "alpha": 1.0, "angle": 0.0},
			},
		})

		// 5. Arrow animation
		if sc.Arrow || idx%2 == 1 {
			tArrow := math.Min(0.45, shotLen*0.7)
			layers = append(layers, map[string]any{
				"type":  "arrow",
				"x":     380,
				"y":     1490,
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
