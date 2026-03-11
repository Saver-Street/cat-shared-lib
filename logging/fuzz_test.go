package logging

import "testing"

func FuzzParseLevel(f *testing.F) {
	f.Add("info")
	f.Add("debug")
	f.Add("warn")
	f.Add("error")
	f.Add("INFO")
	f.Add("")
	f.Add("unknown")

	f.Fuzz(func(t *testing.T, s string) {
		l := ParseLevel(s)
		_ = l.String()
	})
}
