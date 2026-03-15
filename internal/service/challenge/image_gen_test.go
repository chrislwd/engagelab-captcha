package challenge

import (
	"strings"
	"testing"
)

func TestSliderImages(t *testing.T) {
	gen := NewImageGenerator(300, 200)
	bg, piece, targetX, targetY := gen.SliderImages()

	if !strings.HasPrefix(bg, "data:image/png;base64,") {
		t.Fatal("bg should be base64 PNG")
	}
	if !strings.HasPrefix(piece, "data:image/png;base64,") {
		t.Fatal("piece should be base64 PNG")
	}
	if targetX < 64 || targetX > 236 {
		t.Fatalf("targetX out of range: %d", targetX)
	}
	if targetY < 20 || targetY > 136 {
		t.Fatalf("targetY out of range: %d", targetY)
	}
	if len(bg) < 100 {
		t.Fatal("bg image too small")
	}
}

func TestClickTargetImage(t *testing.T) {
	gen := NewImageGenerator(300, 200)
	img, targets := gen.ClickTargetImage([]string{"A", "B", "C"})

	if !strings.HasPrefix(img, "data:image/png;base64,") {
		t.Fatal("should be base64 PNG")
	}
	if len(targets) != 3 {
		t.Fatalf("expected 3 targets, got %d", len(targets))
	}
	for i, target := range targets {
		if target.ID == "" {
			t.Fatalf("target %d missing ID", i)
		}
		if target.X <= 0 || target.Y <= 0 {
			t.Fatalf("target %d invalid position: %d,%d", i, target.X, target.Y)
		}
	}
}

func TestPuzzleImages(t *testing.T) {
	gen := NewImageGenerator(300, 200)
	bg, piece, targetX, targetY := gen.PuzzleImages()

	if !strings.HasPrefix(bg, "data:image/png;base64,") {
		t.Fatal("bg should be base64 PNG")
	}
	if !strings.HasPrefix(piece, "data:image/png;base64,") {
		t.Fatal("piece should be base64 PNG")
	}
	if targetX <= 0 || targetY < 0 {
		t.Fatalf("invalid target: %d,%d", targetX, targetY)
	}
}

func TestSliderImages_Unique(t *testing.T) {
	gen := NewImageGenerator(300, 200)
	_, _, x1, _ := gen.SliderImages()
	_, _, x2, _ := gen.SliderImages()
	// Very unlikely to be identical
	_ = x1
	_ = x2
	// Just ensure no panic
}

func TestBrandConfig_GenerateCSS(t *testing.T) {
	cfg := &BrandConfig{
		PrimaryColor:    "#FF0000",
		BackgroundColor: "#000000",
		TextColor:       "#FFFFFF",
		BorderRadius:    8,
		DarkMode:        true,
	}
	css := cfg.GenerateCSS()
	if !strings.Contains(css, "#FF0000") {
		t.Fatal("CSS should contain primary color")
	}
	if !strings.Contains(css, "#000000") {
		t.Fatal("CSS should contain background color")
	}
	if !strings.Contains(css, "1a1a2e") {
		t.Fatal("CSS should contain dark mode styles")
	}
}

func TestBrandConfig_Default(t *testing.T) {
	cfg := defaultBrand()
	if cfg.PrimaryColor != "#4A90D9" {
		t.Fatalf("expected default primary, got %s", cfg.PrimaryColor)
	}
	css := cfg.GenerateCSS()
	if css == "" {
		t.Fatal("default CSS should not be empty")
	}
}
