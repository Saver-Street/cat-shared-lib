package featureflags_test

import (
	"fmt"

	"github.com/Saver-Street/cat-shared-lib/featureflags"
)

func ExampleManager_Enabled() {
	m := featureflags.NewManager(featureflags.Config{})
	m.Register("dark-mode", "true", "Enable dark mode UI")
	m.Register("beta-api", "false", "Enable beta API endpoints")

	fmt.Println(m.Enabled("dark-mode"))
	fmt.Println(m.Enabled("beta-api"))
	fmt.Println(m.Disabled("beta-api"))
	// Output:
	// true
	// false
	// true
}

func ExampleManager_Value() {
	m := featureflags.NewManager(featureflags.Config{})
	m.Register("max-retries", "5", "Maximum retry attempts")

	fmt.Println(m.Value("max-retries"))
	fmt.Println(m.IntValue("max-retries", 3))
	// Output:
	// 5
	// 5
}
