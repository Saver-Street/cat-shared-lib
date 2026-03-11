package validation

import "testing"

func TestVINValid(t *testing.T) {
	t.Parallel()
	cases := []string{
		"1M8GDM9AXKP042788", // GM vehicle
		"11111111111111111", // all 1s (check digit = 1)
	}
	for _, vin := range cases {
		if err := VIN("vin", vin); err != nil {
			t.Errorf("VIN(%q) = %v, want nil", vin, err)
		}
	}
}

func TestVINValidLowercase(t *testing.T) {
	t.Parallel()
	if err := VIN("vin", "1m8gdm9axkp042788"); err != nil {
		t.Errorf("VIN lowercase should be valid: %v", err)
	}
}

func TestVINValidWithSpaces(t *testing.T) {
	t.Parallel()
	if err := VIN("vin", "  1M8GDM9AXKP042788  "); err != nil {
		t.Errorf("VIN with spaces should be valid: %v", err)
	}
}

func TestVINInvalidLength(t *testing.T) {
	t.Parallel()
	if err := VIN("vin", "1M8GDM9AXK"); err == nil {
		t.Error("short VIN should fail")
	}
}

func TestVINInvalidLetterI(t *testing.T) {
	t.Parallel()
	if err := VIN("vin", "1M8GDM9AIKP042788"); err == nil {
		t.Error("VIN with I should fail")
	}
}

func TestVINInvalidLetterO(t *testing.T) {
	t.Parallel()
	if err := VIN("vin", "1M8GDM9AOKP042788"); err == nil {
		t.Error("VIN with O should fail")
	}
}

func TestVINInvalidLetterQ(t *testing.T) {
	t.Parallel()
	if err := VIN("vin", "1M8GDM9AQKP042788"); err == nil {
		t.Error("VIN with Q should fail")
	}
}

func TestVINInvalidSpecialChar(t *testing.T) {
	t.Parallel()
	if err := VIN("vin", "1M8GDM9A-KP042788"); err == nil {
		t.Error("VIN with special character should fail")
	}
}

func TestVINInvalidCheckDigit(t *testing.T) {
	t.Parallel()
	// Change check digit from X to 0
	if err := VIN("vin", "1M8GDM9A0KP042788"); err == nil {
		t.Error("VIN with wrong check digit should fail")
	}
}

func TestVINCheckDigitX(t *testing.T) {
	t.Parallel()
	// 1M8GDM9AXKP042788 has check digit X (remainder = 10)
	if err := VIN("vin", "1M8GDM9AXKP042788"); err != nil {
		t.Errorf("VIN with X check digit: %v", err)
	}
}

func TestVINErrorField(t *testing.T) {
	t.Parallel()
	err := VIN("vehicle_id", "invalid")
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Field != "vehicle_id" {
		t.Errorf("Field = %q, want %q", err.Field, "vehicle_id")
	}
}

func BenchmarkVIN(b *testing.B) {
	for b.Loop() {
		VIN("vin", "1M8GDM9AXKP042788")
	}
}

func FuzzVIN(f *testing.F) {
	f.Add("1M8GDM9AXKP042788")
	f.Add("11111111111111111")
	f.Add("short")
	f.Add("")
	f.Fuzz(func(t *testing.T, s string) {
		VIN("vin", s)
	})
}
