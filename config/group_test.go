package config

import (
	"testing"

	"github.com/Saver-Street/cat-shared-lib/testkit"
)

func TestLookup_Set(t *testing.T) {
	t.Setenv("TEST_LOOKUP_SET", "hello")
	v, ok := Lookup("TEST_LOOKUP_SET")
	testkit.AssertTrue(t, ok)
	testkit.AssertEqual(t, v, "hello")
}

func TestLookup_Unset(t *testing.T) {
	v, ok := Lookup("TEST_LOOKUP_UNSET_NEVER_SET_12345")
	testkit.AssertTrue(t, !ok)
	testkit.AssertEqual(t, v, "")
}

func TestLookup_EmptyValue(t *testing.T) {
	t.Setenv("TEST_LOOKUP_EMPTY", "")
	v, ok := Lookup("TEST_LOOKUP_EMPTY")
	testkit.AssertTrue(t, !ok)
	testkit.AssertEqual(t, v, "")
}

func TestValidateAll_AllPresent(t *testing.T) {
	t.Setenv("VA_HOST", "localhost")
	t.Setenv("VA_PORT", "5432")
	testkit.AssertNil(t, ValidateAll("VA_HOST", "VA_PORT"))
}

func TestValidateAll_SomeMissing(t *testing.T) {
	t.Setenv("VA_HOST2", "localhost")
	err := ValidateAll("VA_HOST2", "VA_PORT_MISSING_123")
	testkit.AssertNotNil(t, err)
	testkit.AssertContains(t, err.Error(), "VA_PORT_MISSING_123")
}

func TestValidateAll_AllMissing(t *testing.T) {
	err := ValidateAll("VA_MISS_A_123", "VA_MISS_B_123")
	testkit.AssertNotNil(t, err)
	testkit.AssertContains(t, err.Error(), "VA_MISS_A_123")
	testkit.AssertContains(t, err.Error(), "VA_MISS_B_123")
}

func TestValidateAll_Empty(t *testing.T) {
	testkit.AssertNil(t, ValidateAll())
}

func TestValidateAny_OnePresent(t *testing.T) {
	t.Setenv("VANY_KEY", "val")
	testkit.AssertNil(t, ValidateAny("VANY_MISS_123", "VANY_KEY"))
}

func TestValidateAny_NonePresent(t *testing.T) {
	err := ValidateAny("VANY_MISS_X_123", "VANY_MISS_Y_123")
	testkit.AssertNotNil(t, err)
	testkit.AssertContains(t, err.Error(), "at least one")
}

func TestValidateAny_Empty(t *testing.T) {
	err := ValidateAny()
	testkit.AssertNotNil(t, err)
}

func TestFeatureEnabled_True(t *testing.T) {
	for _, v := range []string{"1", "true", "TRUE", "yes", "YES", "on", "ON", "True", "On"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("FE_TEST", v)
			testkit.AssertTrue(t, FeatureEnabled("FE_TEST"))
		})
	}
}

func TestFeatureEnabled_False(t *testing.T) {
	for _, v := range []string{"0", "false", "no", "off", "", "maybe"} {
		t.Run(v, func(t *testing.T) {
			t.Setenv("FE_TEST", v)
			testkit.AssertTrue(t, !FeatureEnabled("FE_TEST"))
		})
	}
}

func TestFeatureEnabled_Unset(t *testing.T) {
	testkit.AssertTrue(t, !FeatureEnabled("FE_NEVER_SET_12345"))
}

func BenchmarkLookup(b *testing.B) {
	b.Setenv("BENCH_LOOKUP", "value")
	for b.Loop() {
		Lookup("BENCH_LOOKUP")
	}
}

func BenchmarkFeatureEnabled(b *testing.B) {
	b.Setenv("BENCH_FE", "true")
	for b.Loop() {
		FeatureEnabled("BENCH_FE")
	}
}
