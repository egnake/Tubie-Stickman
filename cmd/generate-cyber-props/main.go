package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"strconv"
)

// Simple 5x7 font representation for printable ASCII (32..126)
var basicFont5x7 = map[rune][5]byte{
	' ': {0x00, 0x00, 0x00, 0x00, 0x00},
	'!': {0x00, 0x00, 0x5F, 0x00, 0x00},
	'"': {0x00, 0x07, 0x00, 0x07, 0x00},
	'#': {0x14, 0x7F, 0x14, 0x7F, 0x14},
	'$': {0x24, 0x2A, 0x7F, 0x2A, 0x12},
	'%': {0x23, 0x13, 0x08, 0x64, 0x62},
	'&': {0x36, 0x49, 0x55, 0x22, 0x50},
	'\'': {0x00, 0x05, 0x03, 0x00, 0x00},
	'(': {0x00, 0x1C, 0x22, 0x41, 0x00},
	')': {0x00, 0x41, 0x22, 0x1C, 0x00},
	'*': {0x08, 0x2A, 0x1C, 0x2A, 0x08},
	'+': {0x08, 0x08, 0x3E, 0x08, 0x08},
	',': {0x00, 0x50, 0x30, 0x00, 0x00},
	'-': {0x08, 0x08, 0x08, 0x08, 0x08},
	'.': {0x00, 0x60, 0x60, 0x00, 0x00},
	'/': {0x20, 0x10, 0x08, 0x04, 0x02},
	'0': {0x3E, 0x51, 0x49, 0x45, 0x3E},
	'1': {0x00, 0x42, 0x7F, 0x40, 0x00},
	'2': {0x42, 0x61, 0x51, 0x49, 0x46},
	'3': {0x21, 0x41, 0x45, 0x4B, 0x31},
	'4': {0x18, 0x14, 0x12, 0x7F, 0x10},
	'5': {0x27, 0x45, 0x45, 0x45, 0x39},
	'6': {0x3C, 0x4A, 0x49, 0x49, 0x30},
	'7': {0x01, 0x71, 0x09, 0x05, 0x03},
	'8': {0x36, 0x49, 0x49, 0x49, 0x36},
	'9': {0x06, 0x49, 0x49, 0x29, 0x1E},
	':': {0x00, 0x36, 0x36, 0x00, 0x00},
	';': {0x00, 0x56, 0x36, 0x00, 0x00},
	'<': {0x08, 0x14, 0x22, 0x41, 0x00},
	'=': {0x14, 0x14, 0x14, 0x14, 0x14},
	'>': {0x00, 0x41, 0x22, 0x14, 0x08},
	'?': {0x02, 0x01, 0x51, 0x09, 0x06},
	'@': {0x32, 0x49, 0x79, 0x41, 0x3E},
	'A': {0x7E, 0x11, 0x11, 0x11, 0x7E},
	'B': {0x7F, 0x49, 0x49, 0x49, 0x36},
	'C': {0x3E, 0x41, 0x41, 0x41, 0x22},
	'D': {0x7F, 0x41, 0x41, 0x22, 0x1C},
	'E': {0x7F, 0x49, 0x49, 0x49, 0x41},
	'F': {0x7F, 0x09, 0x09, 0x09, 0x01},
	'G': {0x3E, 0x41, 0x49, 0x49, 0x7A},
	'H': {0x7F, 0x08, 0x08, 0x08, 0x7F},
	'I': {0x00, 0x41, 0x7F, 0x41, 0x00},
	'J': {0x20, 0x40, 0x41, 0x3F, 0x01},
	'K': {0x7F, 0x08, 0x14, 0x22, 0x41},
	'L': {0x7F, 0x40, 0x40, 0x40, 0x40},
	'M': {0x7F, 0x02, 0x04, 0x02, 0x7F},
	'N': {0x7F, 0x04, 0x08, 0x10, 0x7F},
	'O': {0x3E, 0x41, 0x41, 0x41, 0x3E},
	'P': {0x7F, 0x09, 0x09, 0x09, 0x06},
	'Q': {0x3E, 0x41, 0x51, 0x21, 0x5E},
	'R': {0x7F, 0x09, 0x19, 0x29, 0x46},
	'S': {0x46, 0x49, 0x49, 0x49, 0x31},
	'T': {0x01, 0x01, 0x7F, 0x01, 0x01},
	'U': {0x3F, 0x40, 0x40, 0x40, 0x3F},
	'V': {0x1F, 0x20, 0x40, 0x20, 0x1F},
	'W': {0x7F, 0x20, 0x18, 0x20, 0x7F},
	'X': {0x63, 0x14, 0x08, 0x14, 0x63},
	'Y': {0x07, 0x08, 0x70, 0x08, 0x07},
	'Z': {0x61, 0x51, 0x49, 0x45, 0x43},
	'[': {0x00, 0x7F, 0x41, 0x41, 0x00},
	'\\': {0x02, 0x04, 0x08, 0x10, 0x20},
	']': {0x00, 0x41, 0x41, 0x7F, 0x00},
	'^': {0x04, 0x02, 0x01, 0x02, 0x04},
	'_': {0x40, 0x40, 0x40, 0x40, 0x40},
	'`': {0x00, 0x01, 0x02, 0x04, 0x00},
	'a': {0x20, 0x54, 0x54, 0x54, 0x78},
	'b': {0x7F, 0x48, 0x44, 0x44, 0x38},
	'c': {0x38, 0x44, 0x44, 0x44, 0x20},
	'd': {0x38, 0x44, 0x44, 0x48, 0x7F},
	'e': {0x38, 0x54, 0x54, 0x54, 0x18},
	'f': {0x08, 0x7E, 0x09, 0x01, 0x02},
	'g': {0x0C, 0x52, 0x52, 0x52, 0x3E},
	'h': {0x7F, 0x08, 0x04, 0x04, 0x78},
	'i': {0x00, 0x44, 0x7D, 0x40, 0x00},
	'j': {0x20, 0x40, 0x44, 0x3D, 0x00},
	'k': {0x7F, 0x10, 0x28, 0x44, 0x00},
	'l': {0x00, 0x41, 0x7F, 0x40, 0x00},
	'm': {0x7C, 0x04, 0x18, 0x04, 0x78},
	'n': {0x7C, 0x08, 0x04, 0x04, 0x78},
	'o': {0x38, 0x44, 0x44, 0x44, 0x38},
	'p': {0x7C, 0x14, 0x14, 0x14, 0x08},
	'q': {0x08, 0x14, 0x14, 0x18, 0x7C},
	'r': {0x7C, 0x08, 0x04, 0x04, 0x08},
	's': {0x48, 0x54, 0x54, 0x54, 0x20},
	't': {0x04, 0x3F, 0x44, 0x40, 0x20},
	'u': {0x3C, 0x40, 0x40, 0x20, 0x7C},
	'v': {0x1C, 0x20, 0x40, 0x20, 0x1C},
	'w': {0x3C, 0x40, 0x30, 0x40, 0x3C},
	'x': {0x44, 0x28, 0x10, 0x28, 0x44},
	'y': {0x0C, 0x50, 0x50, 0x50, 0x3C},
	'z': {0x44, 0x64, 0x54, 0x4C, 0x44},
	'~': {0x04, 0x02, 0x04, 0x08, 0x04},
}

func parseHexColor(s string) color.RGBA {
	if len(s) > 0 && s[0] == '#' {
		s = s[1:]
	}
	if len(s) == 6 {
		val, _ := strconv.ParseUint(s, 16, 32)
		return color.RGBA{R: byte(val >> 16), G: byte((val >> 8) & 0xFF), B: byte(val & 0xFF), A: 255}
	}
	if len(s) == 8 {
		val, _ := strconv.ParseUint(s, 16, 32)
		return color.RGBA{R: byte(val >> 24), G: byte((val >> 16) & 0xFF), B: byte((val >> 8) & 0xFF), A: byte(val & 0xFF)}
	}
	return color.RGBA{0, 0, 0, 255}
}

type Canvas struct {
	img *image.RGBA
	w   int
	h   int
}

func NewCanvas(w, h int) *Canvas {
	return &Canvas{
		img: image.NewRGBA(image.Rect(0, 0, w, h)),
		w:   w,
		h:   h,
	}
}

func (c *Canvas) Set(x, y int, col color.Color) {
	if x >= 0 && x < c.w && y >= 0 && y < c.h {
		c.img.Set(x, y, col)
	}
}

func (c *Canvas) FillRect(x0, y0, x1, y1 int, col color.Color) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			c.Set(x, y, col)
		}
	}
}

func (c *Canvas) DrawLine(x0, y0, x1, y1 int, col color.Color, width int) {
	dx := float64(x1 - x0)
	dy := float64(y1 - y0)
	dist := math.Hypot(dx, dy)
	if dist == 0 {
		c.FillCircle(x0, y0, width/2, col)
		return
	}
	steps := int(dist * 2)
	half := float64(width) / 2.0
	for i := 0; i <= steps; i++ {
		t := float64(i) / float64(steps)
		cx := float64(x0) + t*dx
		cy := float64(y0) + t*dy
		c.FillCircle(int(math.Round(cx)), int(math.Round(cy)), int(math.Round(half)), col)
	}
}

func (c *Canvas) FillCircle(cx, cy, r int, col color.Color) {
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			if (x-cx)*(x-cx)+(y-cy)*(y-cy) <= r*r {
				c.Set(x, y, col)
			}
		}
	}
}

func (c *Canvas) DrawRoundedRect(x0, y0, x1, y1, radius int, fill, outline color.Color, borderWidth int) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	r := float64(radius)
	w := float64(borderWidth)

	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			fx := float64(x)
			fy := float64(y)
			inside := false
			onBorder := false

			// Check inner box
			cx := math.Max(float64(x0)+r, math.Min(fx, float64(x1)-r))
			cy := math.Max(float64(y0)+r, math.Min(fy, float64(y1)-r))
			d := math.Hypot(fx-cx, fy-cy)

			if d <= r {
				inside = true
				if borderWidth > 0 && d >= r-w {
					onBorder = true
				}
			}

			if onBorder && outline != nil {
				c.Set(x, y, outline)
			} else if inside && fill != nil {
				c.Set(x, y, fill)
			}
		}
	}
}

func (c *Canvas) DrawEllipse(x0, y0, x1, y1 int, fill, outline color.Color, borderWidth int) {
	if x0 > x1 {
		x0, x1 = x1, x0
	}
	if y0 > y1 {
		y0, y1 = y1, y0
	}
	cx := float64(x0+x1) / 2.0
	cy := float64(y0+y1) / 2.0
	rx := float64(x1-x0) / 2.0
	ry := float64(y1-y0) / 2.0
	bw := float64(borderWidth)

	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			dx := (float64(x) - cx) / rx
			dy := (float64(y) - cy) / ry
			d := dx*dx + dy*dy
			if d <= 1.0 {
				if borderWidth > 0 {
					innerDx := (float64(x) - cx) / math.Max(1.0, rx-bw)
					innerDy := (float64(y) - cy) / math.Max(1.0, ry-bw)
					innerD := innerDx*innerDx + innerDy*innerDy
					if innerD >= 1.0 && outline != nil {
						c.Set(x, y, outline)
						continue
					}
				}
				if fill != nil {
					c.Set(x, y, fill)
				}
			}
		}
	}
}

func (c *Canvas) DrawArc(x0, y0, x1, y1 int, startDeg, endDeg float64, col color.Color, width int) {
	cx := float64(x0+x1) / 2.0
	cy := float64(y0+y1) / 2.0
	rx := float64(x1-x0) / 2.0
	ry := float64(y1-y0) / 2.0

	// PIL arc starts at 3 o'clock and rotates clockwise.
	// startDeg to endDeg clockwise
	startRad := startDeg * math.Pi / 180.0
	endRad := endDeg * math.Pi / 180.0
	if endRad < startRad {
		endRad += 2 * math.Pi
	}

	steps := int(math.Max(rx, ry) * 4)
	halfW := float64(width) / 2.0
	for i := 0; i <= steps; i++ {
		angle := startRad + float64(i)*(endRad-startRad)/float64(steps)
		px := cx + rx*math.Cos(angle)
		py := cy + ry*math.Sin(angle)
		c.FillCircle(int(math.Round(px)), int(math.Round(py)), int(math.Round(halfW)), col)
	}
}

func (c *Canvas) DrawPolygon(pts [][2]int, fill, outline color.Color, borderWidth int) {
	n := len(pts)
	if n < 3 {
		return
	}
	if fill != nil {
		minY, maxY := pts[0][1], pts[0][1]
		for _, p := range pts {
			if p[1] < minY {
				minY = p[1]
			}
			if p[1] > maxY {
				maxY = p[1]
			}
		}
		for y := minY; y <= maxY; y++ {
			var nodeX []int
			j := n - 1
			for i := 0; i < n; i++ {
				yi, yj := pts[i][1], pts[j][1]
				xi, xj := pts[i][0], pts[j][0]
				if (yi < y && yj >= y) || (yj < y && yi >= y) {
					nx := xi + (y-yi)*(xj-xi)/(yj-yi)
					nodeX = append(nodeX, nx)
				}
				j = i
			}
			for i := 0; i < len(nodeX)-1; i++ {
				for k := i + 1; k < len(nodeX); k++ {
					if nodeX[i] > nodeX[k] {
						nodeX[i], nodeX[k] = nodeX[k], nodeX[i]
					}
				}
			}
			for i := 0; i < len(nodeX); i += 2 {
				if i+1 < len(nodeX) {
					c.FillRect(nodeX[i], y, nodeX[i+1], y, fill)
				}
			}
		}
	}
	if outline != nil && borderWidth > 0 {
		for i := 0; i < n; i++ {
			next := (i + 1) % n
			c.DrawLine(pts[i][0], pts[i][1], pts[next][0], pts[next][1], outline, borderWidth)
		}
	}
}

func (c *Canvas) DrawText(x, y int, text string, col color.Color, scale int) {
	if scale <= 0 {
		scale = 2
	}
	curX := x
	for _, ch := range text {
		bitmap, ok := basicFont5x7[ch]
		if !ok {
			bitmap = basicFont5x7['?']
		}
		for colIdx := 0; colIdx < 5; colIdx++ {
			colBits := bitmap[colIdx]
			for rowIdx := 0; rowIdx < 7; rowIdx++ {
				if (colBits & (1 << rowIdx)) != 0 {
					c.FillRect(curX+colIdx*scale, y+rowIdx*scale, curX+(colIdx+1)*scale-1, y+(rowIdx+1)*scale-1, col)
				}
			}
		}
		curX += 6 * scale
	}
}

func savePNG(path string, img image.Image, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("File already exists, skipping: %s\n", path)
			return nil
		}
	}
	_ = os.MkdirAll(filepath.Dir(path), 0755)
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func generateProps(outDir string, force bool) error {
	// 1. Cyber Laptop
	{
		c := NewCanvas(512, 512)
		c.DrawRoundedRect(60, 80, 452, 340, 16, parseHexColor("#111827"), parseHexColor("#374151"), 6)
		c.DrawRoundedRect(76, 96, 436, 324, 8, parseHexColor("#030712"), parseHexColor("#00FF66"), 3)
		c.DrawText(95, 115, "> root@kali:~# ./exploit.sh", parseHexColor("#00FF66"), 2)
		c.DrawText(95, 145, "[*] Scanning target port 443...", parseHexColor("#00F0FF"), 2)
		c.DrawText(95, 175, "[+] VULNERABILITY FOUND: CVE-2026", parseHexColor("#FFE500"), 2)
		c.DrawText(95, 205, "[!] BYPASSING FIREWALL... [OK]", parseHexColor("#FF3B30"), 2)
		c.DrawText(95, 235, ">>> ACCESS GRANTED: ROOT", parseHexColor("#00FF66"), 2)
		c.DrawPolygon([][2]int{{30, 370}, {482, 370}, {500, 395}, {12, 395}}, parseHexColor("#1F2937"), parseHexColor("#4B5563"), 1)
		c.FillRect(210, 375, 302, 390, parseHexColor("#111827"))
		if err := savePNG(filepath.Join(outDir, "cyber-laptop.png"), c.img, force); err != nil {
			return err
		}
	}

	// 2. Cyber Padlock Locked
	{
		c := NewCanvas(512, 512)
		c.DrawArc(156, 70, 356, 290, 180, 360, parseHexColor("#00FF66"), 38)
		c.FillRect(156, 170, 194, 250, parseHexColor("#00FF66"))
		c.FillRect(318, 170, 356, 250, parseHexColor("#00FF66"))
		c.DrawRoundedRect(106, 210, 406, 450, 30, parseHexColor("#0A1910"), parseHexColor("#00FF66"), 10)
		c.DrawEllipse(236, 280, 276, 320, parseHexColor("#00FF66"), nil, 0)
		c.DrawPolygon([][2]int{{244, 310}, {268, 310}, {278, 380}, {234, 380}}, parseHexColor("#00FF66"), nil, 0)
		c.DrawText(170, 400, "AES-256 SECURED", parseHexColor("#00FF66"), 2)
		if err := savePNG(filepath.Join(outDir, "cyber-padlock-locked.png"), c.img, force); err != nil {
			return err
		}
	}

	// 3. Cyber Padlock Unlocked
	{
		c := NewCanvas(512, 512)
		c.DrawArc(196, 30, 396, 250, 180, 360, parseHexColor("#FF3B30"), 38)
		c.FillRect(196, 130, 234, 250, parseHexColor("#FF3B30"))
		c.FillRect(358, 130, 396, 180, parseHexColor("#FF3B30"))
		c.DrawRoundedRect(106, 210, 406, 450, 30, parseHexColor("#200A0A"), parseHexColor("#FF3B30"), 10)
		c.DrawEllipse(236, 280, 276, 320, parseHexColor("#FF3B30"), nil, 0)
		c.DrawPolygon([][2]int{{244, 310}, {268, 310}, {278, 380}, {234, 380}}, parseHexColor("#FF3B30"), nil, 0)
		c.DrawText(160, 400, "PASSWORD CRACKED!", parseHexColor("#FF3B30"), 2)
		if err := savePNG(filepath.Join(outDir, "cyber-padlock-unlocked.png"), c.img, force); err != nil {
			return err
		}
	}

	// 4. Cyber Shield
	{
		c := NewCanvas(512, 512)
		pts := [][2]int{{256, 50}, {430, 90}, {410, 320}, {256, 460}, {102, 320}, {82, 90}}
		c.DrawPolygon(pts, parseHexColor("#061A24"), parseHexColor("#00E5FF"), 10)
		c.DrawLine(180, 260, 235, 320, parseHexColor("#00FF66"), 22)
		c.DrawLine(235, 320, 345, 180, parseHexColor("#00FF66"), 22)
		c.DrawText(180, 370, "FIREWALL: ON", parseHexColor("#00E5FF"), 2)
		if err := savePNG(filepath.Join(outDir, "cyber-shield.png"), c.img, force); err != nil {
			return err
		}
	}

	// 5. Cyber Skull
	{
		c := NewCanvas(512, 512)
		c.DrawEllipse(136, 80, 376, 320, parseHexColor("#1E0B0B"), parseHexColor("#FF3B30"), 10)
		c.DrawRoundedRect(186, 290, 326, 400, 16, parseHexColor("#1E0B0B"), parseHexColor("#FF3B30"), 10)
		c.DrawEllipse(176, 170, 236, 240, parseHexColor("#FF3B30"), nil, 0)
		c.DrawEllipse(276, 170, 336, 240, parseHexColor("#FF3B30"), nil, 0)
		c.DrawPolygon([][2]int{{256, 260}, {242, 290}, {270, 290}}, parseHexColor("#FF3B30"), nil, 0)
		for _, x := range []int{216, 246, 276, 296} {
			c.DrawLine(x, 320, x, 380, parseHexColor("#FF3B30"), 6)
		}
		c.DrawText(150, 430, "CRITICAL THREAT!", parseHexColor("#FF3B30"), 2)
		if err := savePNG(filepath.Join(outDir, "cyber-skull.png"), c.img, force); err != nil {
			return err
		}
	}

	// 6. Cyber Database
	{
		c := NewCanvas(512, 512)
		c.DrawEllipse(106, 60, 406, 150, parseHexColor("#1E293B"), parseHexColor("#38BDF8"), 8)
		c.FillRect(106, 105, 406, 380, parseHexColor("#0F172A"))
		c.DrawLine(106, 105, 106, 380, parseHexColor("#38BDF8"), 8)
		c.DrawLine(406, 105, 406, 380, parseHexColor("#38BDF8"), 8)
		c.DrawArc(106, 160, 406, 250, 0, 180, parseHexColor("#38BDF8"), 8)
		c.DrawArc(106, 270, 406, 360, 0, 180, parseHexColor("#38BDF8"), 8)
		c.DrawEllipse(106, 340, 406, 430, parseHexColor("#1E293B"), parseHexColor("#38BDF8"), 8)
		c.DrawLine(256, 180, 230, 230, parseHexColor("#EF4444"), 8)
		c.DrawLine(230, 230, 290, 280, parseHexColor("#EF4444"), 8)
		c.DrawLine(290, 280, 240, 350, parseHexColor("#EF4444"), 8)
		c.DrawText(140, 200, "' OR 1=1 --", parseHexColor("#EF4444"), 2)
		c.DrawText(160, 450, "SQL INJECTION!", parseHexColor("#EF4444"), 2)
		if err := savePNG(filepath.Join(outDir, "cyber-database.png"), c.img, force); err != nil {
			return err
		}
	}

	// 7. Cyber Wi-Fi
	{
		c := NewCanvas(512, 512)
		c.DrawEllipse(236, 380, 276, 420, parseHexColor("#F59E0B"), nil, 0)
		c.DrawArc(186, 300, 326, 440, 210, 330, parseHexColor("#F59E0B"), 16)
		c.DrawArc(136, 220, 376, 460, 210, 330, parseHexColor("#F59E0B"), 16)
		c.DrawArc(86, 140, 426, 480, 210, 330, parseHexColor("#EF4444"), 20)
		c.DrawText(140, 90, "EVIL TWIN / ROGUE AP", parseHexColor("#EF4444"), 2)
		c.DrawText(170, 440, "UNSECURED WI-FI", parseHexColor("#F59E0B"), 2)
		if err := savePNG(filepath.Join(outDir, "cyber-wifi.png"), c.img, force); err != nil {
			return err
		}
	}

	// 8. Cyber Bug
	{
		c := NewCanvas(512, 512)
		c.DrawEllipse(156, 160, 356, 380, parseHexColor("#2B1010"), parseHexColor("#EF4444"), 10)
		c.DrawEllipse(186, 100, 326, 210, parseHexColor("#2B1010"), parseHexColor("#EF4444"), 8)
		c.DrawLine(220, 120, 170, 50, parseHexColor("#EF4444"), 8)
		c.DrawLine(292, 120, 342, 50, parseHexColor("#EF4444"), 8)
		for _, pair := range [][2]int{{220, 110}, {280, 120}, {340, 100}} {
			y, dx := pair[0], pair[1]
			c.DrawLine(156, y, 156-dx, y-20, parseHexColor("#EF4444"), 8)
			c.DrawLine(156-dx, y-20, 156-dx-30, y+30, parseHexColor("#EF4444"), 8)
			c.DrawLine(356, y, 356+dx, y-20, parseHexColor("#EF4444"), 8)
			c.DrawLine(356+dx, y-20, 356+dx+30, y+30, parseHexColor("#EF4444"), 8)
		}
		c.DrawText(160, 420, "ZERO-DAY EXPLOIT", parseHexColor("#EF4444"), 2)
		if err := savePNG(filepath.Join(outDir, "cyber-bug.png"), c.img, force); err != nil {
			return err
		}
	}

	// 9. Cyber Terminal
	{
		c := NewCanvas(512, 512)
		c.DrawRoundedRect(40, 60, 472, 452, 16, parseHexColor("#0D1117"), parseHexColor("#30363D"), 6)
		c.DrawRoundedRect(40, 60, 472, 110, 16, parseHexColor("#161B22"), nil, 0)
		c.FillRect(40, 85, 472, 110, parseHexColor("#161B22"))
		c.DrawEllipse(65, 78, 85, 98, parseHexColor("#FF5F56"), nil, 0)
		c.DrawEllipse(95, 78, 115, 98, parseHexColor("#FFBD2E"), nil, 0)
		c.DrawEllipse(125, 78, 145, 98, parseHexColor("#27C93F"), nil, 0)
		c.DrawText(190, 77, "bash - 80x24", parseHexColor("#8B949E"), 2)
		c.DrawText(65, 130, "$ nmap -sV 192.168.1.1", parseHexColor("#00FF66"), 2)
		c.DrawText(65, 170, "Starting Nmap 7.94 scan...", parseHexColor("#E6EDF3"), 2)
		c.DrawText(65, 210, "PORT   STATE SERVICE VERSION", parseHexColor("#7EE787"), 2)
		c.DrawText(65, 250, "22/tcp open  ssh     OpenSSH 8.2", parseHexColor("#E6EDF3"), 2)
		c.DrawText(65, 290, "80/tcp open  http    Apache 2.4", parseHexColor("#E6EDF3"), 2)
		c.DrawText(65, 330, "[!] REVERSE SHELL OPENED!", parseHexColor("#FF7B72"), 2)
		c.DrawText(65, 370, "root@pwned:~# _", parseHexColor("#58A6FF"), 2)
		if err := savePNG(filepath.Join(outDir, "cyber-terminal.png"), c.img, force); err != nil {
			return err
		}
	}

	fmt.Printf("All 9 Cyber Props successfully generated in %s\n", outDir)
	return nil
}

func main() {
	var outputDir string
	var force bool

	flag.StringVar(&outputDir, "output-dir", filepath.Join("skills", "directing-stickman-videos", "assets", "b1"), "Directory to save generated cyber props")
	flag.BoolVar(&force, "force", false, "Overwrite existing files if present")
	flag.Parse()

	if err := generateProps(outputDir, force); err != nil {
		fmt.Fprintf(os.Stderr, "Error generating cyber props: %v\n", err)
		os.Exit(1)
	}
}

// Suppress unused import warning for draw
var _ = draw.Src
