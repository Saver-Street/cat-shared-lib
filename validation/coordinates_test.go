package validation

import (
	"math"
	"testing"
)

func TestLatitude(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		ok    bool
	}{
		{"zero", 0, true},
		{"min", -90, true},
		{"max", 90, true},
		{"positive", 40.7128, true},
		{"negative", -33.8688, true},
		{"too low", -90.1, false},
		{"too high", 90.1, false},
		{"NaN", math.NaN(), false},
		{"Inf", math.Inf(1), false},
		{"neg Inf", math.Inf(-1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Latitude("lat", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("Latitude(%v) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("Latitude(%v) = nil, want error", tt.value)
			}
		})
	}
}

func TestLongitude(t *testing.T) {
	tests := []struct {
		name  string
		value float64
		ok    bool
	}{
		{"zero", 0, true},
		{"min", -180, true},
		{"max", 180, true},
		{"positive", 74.0060, true},
		{"negative", -151.2093, true},
		{"too low", -180.1, false},
		{"too high", 180.1, false},
		{"NaN", math.NaN(), false},
		{"Inf", math.Inf(1), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Longitude("lon", tt.value)
			if tt.ok && err != nil {
				t.Fatalf("Longitude(%v) = %v, want nil", tt.value, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("Longitude(%v) = nil, want error", tt.value)
			}
		})
	}
}

func TestCoordinates(t *testing.T) {
	tests := []struct {
		name string
		lat  float64
		lon  float64
		ok   bool
	}{
		{"New York", 40.7128, -74.0060, true},
		{"Sydney", -33.8688, 151.2093, true},
		{"origin", 0, 0, true},
		{"bad lat", 91, 0, false},
		{"bad lon", 0, 181, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Coordinates("lat", tt.lat, "lon", tt.lon)
			if tt.ok && err != nil {
				t.Fatalf("Coordinates(%v, %v) = %v, want nil", tt.lat, tt.lon, err)
			}
			if !tt.ok && err == nil {
				t.Fatalf("Coordinates(%v, %v) = nil, want error", tt.lat, tt.lon)
			}
		})
	}
}

func TestCoordinatesInRadius(t *testing.T) {
	tests := []struct {
		name                       string
		lat, lon, centLat, centLon float64
		radiusKm                   float64
		ok                         bool
	}{
		{"within radius", 40.7128, -74.0060, 40.7580, -73.9855, 10, true},
		{"outside radius", 40.7128, -74.0060, 51.5074, -0.1278, 100, false},
		{"same point", 0, 0, 0, 0, 1, true},
		{"bad lat", 91, 0, 0, 0, 100, false},
		{"bad center lat", 0, 0, 91, 0, 100, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CoordinatesInRadius("pos", tt.lat, tt.lon, tt.centLat, tt.centLon, tt.radiusKm)
			if tt.ok && err != nil {
				t.Fatalf("CoordinatesInRadius = %v, want nil", err)
			}
			if !tt.ok && err == nil {
				t.Fatal("CoordinatesInRadius = nil, want error")
			}
		})
	}
}

func TestHaversine(t *testing.T) {
	// New York to London: ~5570 km
	dist := haversine(40.7128, -74.0060, 51.5074, -0.1278)
	if dist < 5500 || dist > 5600 {
		t.Fatalf("haversine(NYC, London) = %.2f, want ~5570", dist)
	}

	// Same point
	dist = haversine(0, 0, 0, 0)
	if dist != 0 {
		t.Fatalf("haversine(0,0,0,0) = %f, want 0", dist)
	}
}

func TestToRad(t *testing.T) {
	if got := toRad(180); math.Abs(got-math.Pi) > 1e-10 {
		t.Fatalf("toRad(180) = %f, want %f", got, math.Pi)
	}
	if got := toRad(0); got != 0 {
		t.Fatalf("toRad(0) = %f, want 0", got)
	}
}

func BenchmarkLatitude(b *testing.B) {
	for b.Loop() {
		Latitude("lat", 40.7128)
	}
}

func BenchmarkCoordinatesInRadius(b *testing.B) {
	for b.Loop() {
		CoordinatesInRadius("pos", 40.7128, -74.0060, 40.7580, -73.9855, 10)
	}
}

func BenchmarkHaversine(b *testing.B) {
	for b.Loop() {
		haversine(40.7128, -74.0060, 51.5074, -0.1278)
	}
}

func FuzzLatitude(f *testing.F) {
	f.Add(0.0)
	f.Add(90.0)
	f.Add(-90.0)
	f.Add(91.0)
	f.Add(math.NaN())

	f.Fuzz(func(t *testing.T, v float64) {
		_ = Latitude("lat", v)
	})
}

func FuzzLongitude(f *testing.F) {
	f.Add(0.0)
	f.Add(180.0)
	f.Add(-180.0)
	f.Add(181.0)

	f.Fuzz(func(t *testing.T, v float64) {
		_ = Longitude("lon", v)
	})
}
