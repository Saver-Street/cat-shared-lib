package config_test

import (
	"fmt"
	"os"
	"time"

	"github.com/Saver-Street/cat-shared-lib/config"
)

func ExampleString() {
	os.Setenv("APP_NAME", "billing-service")
	defer os.Unsetenv("APP_NAME")

	fmt.Println(config.String("APP_NAME", "default"))
	fmt.Println(config.String("MISSING_KEY", "fallback"))
	// Output:
	// billing-service
	// fallback
}

func ExampleInt() {
	os.Setenv("PORT", "8080")
	defer os.Unsetenv("PORT")

	fmt.Println(config.Int("PORT", 3000))
	fmt.Println(config.Int("MISSING_PORT", 3000))
	// Output:
	// 8080
	// 3000
}

func ExampleBool() {
	os.Setenv("DEBUG", "true")
	defer os.Unsetenv("DEBUG")

	fmt.Println(config.Bool("DEBUG", false))
	fmt.Println(config.Bool("MISSING_FLAG", false))
	// Output:
	// true
	// false
}

func ExampleDuration() {
	os.Setenv("TIMEOUT", "30s")
	defer os.Unsetenv("TIMEOUT")

	fmt.Println(config.Duration("TIMEOUT", 5*time.Second))
	fmt.Println(config.Duration("MISSING_DUR", 5*time.Second))
	// Output:
	// 30s
	// 5s
}

func ExampleStringRequired() {
	os.Setenv("DB_HOST", "localhost")
	defer os.Unsetenv("DB_HOST")

	val, err := config.StringRequired("DB_HOST")
	fmt.Println(val, err)

	_, err = config.StringRequired("MISSING_REQUIRED")
	fmt.Println(err)
	// Output:
	// localhost <nil>
	// config: required environment variable MISSING_REQUIRED is not set
}

func ExampleStringSlice() {
	os.Setenv("ALLOWED_ORIGINS", "http://localhost,https://example.com")
	defer os.Unsetenv("ALLOWED_ORIGINS")

	fmt.Println(config.StringSlice("ALLOWED_ORIGINS", nil))
	fmt.Println(config.StringSlice("MISSING_SLICE", []string{"default"}))
	// Output:
	// [http://localhost https://example.com]
	// [default]
}

func ExampleStringMap() {
	os.Setenv("LABELS", "env=prod,region=us-east-1")
	defer os.Unsetenv("LABELS")

	m := config.StringMap("LABELS", nil)
	fmt.Println(m["env"], m["region"])
	// Output:
	// prod us-east-1
}

func ExampleValidate() {
	os.Setenv("REQUIRED_A", "set")
	defer os.Unsetenv("REQUIRED_A")

	err := config.Validate("REQUIRED_A")
	fmt.Println(err)

	err = config.Validate("REQUIRED_A", "MISSING_B")
	fmt.Println(err)
	// Output:
	// <nil>
	// config: missing required environment variables: MISSING_B
}

func ExampleLookup() {
	os.Setenv("MY_KEY", "hello")
	v, ok := config.Lookup("MY_KEY")
	fmt.Println(v, ok)
	os.Unsetenv("MY_KEY")
	// Output:
	// hello true
}

func ExampleFeatureEnabled() {
	os.Setenv("DARK_MODE", "true")
	fmt.Println(config.FeatureEnabled("DARK_MODE"))
	os.Unsetenv("DARK_MODE")
	// Output:
	// true
}
