package challenge

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"bytes"
	"math"
	"math/big"
)

// ImageGenerator creates CAPTCHA challenge images (slider backgrounds, puzzle pieces, click targets).
type ImageGenerator struct {
	width  int
	height int
}

func NewImageGenerator(width, height int) *ImageGenerator {
	if width == 0 { width = 300 }
	if height == 0 { height = 200 }
	return &ImageGenerator{width: width, height: height}
}

// SliderImages generates a background image with a notch and a slider piece image.
// Returns (bgBase64, pieceBase64, targetX, targetY).
func (g *ImageGenerator) SliderImages() (string, string, int, int) {
	pieceSize := 44
	targetX := randInt(pieceSize+20, g.width-pieceSize-20)
	targetY := randInt(20, g.height-pieceSize-20)

	// Generate background with gradient and shapes
	bg := image.NewRGBA(image.Rect(0, 0, g.width, g.height))
	g.drawGradientBg(bg)
	g.drawRandomShapes(bg, 8)

	// Create piece from the background region
	piece := image.NewRGBA(image.Rect(0, 0, pieceSize, pieceSize))
	for y := 0; y < pieceSize; y++ {
		for x := 0; x < pieceSize; x++ {
			// Circle mask
			cx, cy := float64(x)-float64(pieceSize)/2, float64(y)-float64(pieceSize)/2
			if cx*cx+cy*cy <= float64(pieceSize*pieceSize)/4 {
				piece.Set(x, y, bg.At(targetX+x, targetY+y))
			} else {
				piece.Set(x, y, color.Transparent)
			}
		}
	}

	// Draw notch on background (semi-transparent overlay)
	notchColor := color.RGBA{0, 0, 0, 80}
	for y := 0; y < pieceSize; y++ {
		for x := 0; x < pieceSize; x++ {
			cx, cy := float64(x)-float64(pieceSize)/2, float64(y)-float64(pieceSize)/2
			if cx*cx+cy*cy <= float64(pieceSize*pieceSize)/4 {
				bg.Set(targetX+x, targetY+y, notchColor)
			}
		}
	}

	// Draw border on notch
	for angle := 0.0; angle < 360; angle += 1 {
		rad := angle * math.Pi / 180
		bx := int(float64(pieceSize)/2*math.Cos(rad)) + targetX + pieceSize/2
		by := int(float64(pieceSize)/2*math.Sin(rad)) + targetY + pieceSize/2
		if bx >= 0 && bx < g.width && by >= 0 && by < g.height {
			bg.Set(bx, by, color.RGBA{255, 255, 255, 180})
		}
	}

	bgData := encodePNG(bg)
	pieceData := encodePNG(piece)

	return bgData, pieceData, targetX, targetY
}

// ClickTargetImage generates an image with labeled clickable targets.
// Returns (imageBase64, targets with positions).
func (g *ImageGenerator) ClickTargetImage(labels []string) (string, []ClickTarget) {
	img := image.NewRGBA(image.Rect(0, 0, g.width, g.height))
	g.drawGradientBg(img)

	var targets []ClickTarget
	used := make([]image.Rectangle, 0)
	targetSize := 36

	for i, label := range labels {
		// Find non-overlapping position
		var x, y int
		for attempts := 0; attempts < 50; attempts++ {
			x = randInt(targetSize, g.width-targetSize*2)
			y = randInt(targetSize, g.height-targetSize*2)
			r := image.Rect(x, y, x+targetSize, y+targetSize)
			overlap := false
			for _, u := range used {
				if r.Overlaps(u.Inset(-10)) {
					overlap = true
					break
				}
			}
			if !overlap {
				used = append(used, r)
				break
			}
		}

		// Draw target circle
		colors := []color.RGBA{
			{66, 133, 244, 255},  // blue
			{234, 67, 53, 255},   // red
			{251, 188, 4, 255},   // yellow
			{52, 168, 83, 255},   // green
			{171, 71, 188, 255},  // purple
		}
		c := colors[i%len(colors)]

		for dy := -targetSize / 2; dy < targetSize/2; dy++ {
			for dx := -targetSize / 2; dx < targetSize/2; dx++ {
				if dx*dx+dy*dy <= (targetSize/2)*(targetSize/2) {
					px, py := x+targetSize/2+dx, y+targetSize/2+dy
					if px >= 0 && px < g.width && py >= 0 && py < g.height {
						img.Set(px, py, c)
					}
				}
			}
		}

		// Draw label letter in center (simple pixel font for first char)
		g.drawChar(img, x+targetSize/2-3, y+targetSize/2-4, label[0], color.White)

		targets = append(targets, ClickTarget{
			ID:    fmt.Sprintf("t_%d", i),
			Label: label,
			X:     x + targetSize/2,
			Y:     y + targetSize/2,
		})
	}

	return encodePNG(img), targets
}

type ClickTarget struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	X     int    `json:"x"`
	Y     int    `json:"y"`
}

// PuzzleImages generates a background with a puzzle-shaped cutout and matching piece.
func (g *ImageGenerator) PuzzleImages() (string, string, int, int) {
	pieceW, pieceH := 50, 50
	targetX := randInt(pieceW+10, g.width-pieceW-10)
	targetY := randInt(10, g.height-pieceH-10)

	bg := image.NewRGBA(image.Rect(0, 0, g.width, g.height))
	g.drawGradientBg(bg)
	g.drawRandomShapes(bg, 6)

	// Create puzzle piece (rectangle with tab)
	piece := image.NewRGBA(image.Rect(0, 0, pieceW+10, pieceH))

	for y := 0; y < pieceH; y++ {
		for x := 0; x < pieceW; x++ {
			piece.Set(x, y, bg.At(targetX+x, targetY+y))
		}
	}
	// Tab protrusion
	tabR := 8
	tabCX, tabCY := pieceW, pieceH/2
	for dy := -tabR; dy <= tabR; dy++ {
		for dx := -tabR; dx <= tabR; dx++ {
			if dx*dx+dy*dy <= tabR*tabR {
				sx := targetX + tabCX + dx
				sy := targetY + tabCY + dy
				if sx >= 0 && sx < g.width && sy >= 0 && sy < g.height {
					piece.Set(tabCX+dx, tabCY+dy, bg.At(sx, sy))
				}
			}
		}
	}

	// Draw cutout on bg
	cutColor := color.RGBA{0, 0, 0, 60}
	for y := 0; y < pieceH; y++ {
		for x := 0; x < pieceW; x++ {
			bg.Set(targetX+x, targetY+y, cutColor)
		}
	}
	for dy := -tabR; dy <= tabR; dy++ {
		for dx := -tabR; dx <= tabR; dx++ {
			if dx*dx+dy*dy <= tabR*tabR {
				px := targetX + tabCX + dx
				py := targetY + tabCY + dy
				if px >= 0 && px < g.width && py >= 0 && py < g.height {
					bg.Set(px, py, cutColor)
				}
			}
		}
	}

	return encodePNG(bg), encodePNG(piece), targetX, targetY
}

// --- Drawing helpers ---

func (g *ImageGenerator) drawGradientBg(img *image.RGBA) {
	r1, g1, b1 := randInt(100, 200), randInt(100, 200), randInt(150, 230)
	r2, g2, b2 := randInt(50, 150), randInt(80, 180), randInt(100, 200)

	for y := 0; y < g.height; y++ {
		t := float64(y) / float64(g.height)
		r := uint8(float64(r1)*(1-t) + float64(r2)*t)
		gv := uint8(float64(g1)*(1-t) + float64(g2)*t)
		b := uint8(float64(b1)*(1-t) + float64(b2)*t)
		for x := 0; x < g.width; x++ {
			// Add subtle noise
			noise := randInt(-8, 8)
			img.Set(x, y, color.RGBA{clamp(int(r) + noise), clamp(int(gv) + noise), clamp(int(b) + noise), 255})
		}
	}
}

func (g *ImageGenerator) drawRandomShapes(img *image.RGBA, count int) {
	for i := 0; i < count; i++ {
		cx := randInt(0, g.width)
		cy := randInt(0, g.height)
		radius := randInt(15, 50)
		c := color.RGBA{
			uint8(randInt(50, 200)),
			uint8(randInt(50, 200)),
			uint8(randInt(50, 200)),
			uint8(randInt(40, 100)),
		}
		for dy := -radius; dy <= radius; dy++ {
			for dx := -radius; dx <= radius; dx++ {
				if dx*dx+dy*dy <= radius*radius {
					px, py := cx+dx, cy+dy
					if px >= 0 && px < g.width && py >= 0 && py < g.height {
						existing := img.At(px, py).(color.RGBA)
						img.Set(px, py, blendColor(existing, c))
					}
				}
			}
		}
	}
}

// Simple 5x7 pixel font for single character
func (g *ImageGenerator) drawChar(img *image.RGBA, x, y int, ch byte, c color.Color) {
	// Simplified: draw a small filled rectangle as placeholder for the character
	for dy := 0; dy < 7; dy++ {
		for dx := 0; dx < 5; dx++ {
			img.Set(x+dx, y+dy, c)
		}
	}
}

func encodePNG(img image.Image) string {
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

func blendColor(bg, fg color.RGBA) color.RGBA {
	a := float64(fg.A) / 255
	return color.RGBA{
		uint8(float64(bg.R)*(1-a) + float64(fg.R)*a),
		uint8(float64(bg.G)*(1-a) + float64(fg.G)*a),
		uint8(float64(bg.B)*(1-a) + float64(fg.B)*a),
		255,
	}
}

func clamp(v int) uint8 {
	if v < 0 { return 0 }
	if v > 255 { return 255 }
	return uint8(v)
}

func randInt(min, max int) int {
	if min >= max { return min }
	n, _ := rand.Int(rand.Reader, big.NewInt(int64(max-min)))
	return int(n.Int64()) + min
}
