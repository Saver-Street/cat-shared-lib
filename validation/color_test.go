package validation

import (
	"testing"
)

func TestHexAlphaColor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"4-digit", "#fffa", false},
		{"8-digit", "#a1b2c3ff", false},
		{"8-digit uppercase", "#A1B2C3FF", false},
		{"empty", "", true},
		{"3-digit no alpha", "#fff", true},
		{"6-digit no alpha", "#a1b2c3", true},
		{"invalid char", "#gggggggg", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := HexAlphaColor("color", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("HexAlphaColor(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestRGBColor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid", "rgb(255, 128, 0)", false},
		{"zeros", "rgb(0, 0, 0)", false},
		{"max", "rgb(255, 255, 255)", false},
		{"no spaces", "rgb(10,20,30)", false},
		{"extra spaces", "rgb(  10 ,  20 ,  30 )", false},
		{"r too high", "rgb(256, 0, 0)", true},
		{"g too high", "rgb(0, 256, 0)", true},
		{"b too high", "rgb(0, 0, 256)", true},
		{"negative", "rgb(-1, 0, 0)", true},
		{"decimal", "rgb(1.5, 0, 0)", true},
		{"empty", "", true},
		{"no parens", "rgb 0,0,0", true},
		{"missing component", "rgb(0, 0)", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := RGBColor("color", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("RGBColor(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestHSLColor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"valid", "hsl(120, 50%, 75%)", false},
		{"zeros", "hsl(0, 0%, 0%)", false},
		{"max", "hsl(360, 100%, 100%)", false},
		{"no spaces", "hsl(120,50%,75%)", false},
		{"hue too high", "hsl(361, 50%, 75%)", true},
		{"sat too high", "hsl(120, 101%, 75%)", true},
		{"light too high", "hsl(120, 50%, 101%)", true},
		{"no percent", "hsl(120, 50, 75)", true},
		{"empty", "", true},
		{"negative hue", "hsl(-10, 50%, 75%)", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := HSLColor("color", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("HSLColor(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestCSSColor(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"hex 3-digit", "#fff", false},
		{"hex 6-digit", "#a1b2c3", false},
		{"rgb", "rgb(255, 128, 0)", false},
		{"hsl", "hsl(120, 50%, 75%)", false},
		{"named color", "red", false},
		{"named color uppercase", "Red", false},
		{"named color mixed", "DarkSlateGray", false},
		{"transparent", "transparent", false},
		{"invalid rgb values", "rgb(999, 0, 0)", true},
		{"invalid hsl values", "hsl(999, 0%, 0%)", true},
		{"unknown name", "notacolor", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := CSSColor("color", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("CSSColor(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func BenchmarkRGBColor(b *testing.B) {
	for range b.N {
		_ = RGBColor("color", "rgb(255, 128, 0)")
	}
}

func BenchmarkHSLColor(b *testing.B) {
	for range b.N {
		_ = HSLColor("color", "hsl(120, 50%, 75%)")
	}
}

func BenchmarkCSSColor(b *testing.B) {
	for range b.N {
		_ = CSSColor("color", "rebeccapurple")
	}
}

func FuzzRGBColor(f *testing.F) {
	f.Add("rgb(255, 128, 0)")
	f.Add("rgb(0,0,0)")
	f.Add("")
	f.Fuzz(func(t *testing.T, s string) {
		_ = RGBColor("color", s)
	})
}

func FuzzHSLColor(f *testing.F) {
	f.Add("hsl(120, 50%, 75%)")
	f.Add("hsl(0,0%,0%)")
	f.Add("")
	f.Fuzz(func(t *testing.T, s string) {
		_ = HSLColor("color", s)
	})
}
