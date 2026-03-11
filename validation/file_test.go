package validation

import (
	"testing"
)

func TestFileExtension(t *testing.T) {
	t.Parallel()
	allowed := []string{".jpg", ".png", ".gif"}
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"jpg", "photo.jpg", false},
		{"png", "image.png", false},
		{"gif", "animation.gif", false},
		{"case insensitive", "PHOTO.JPG", false},
		{"with path", "/uploads/photo.jpg", false},
		{"not allowed", "doc.pdf", true},
		{"no extension", "readme", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := FileExtension("file", tt.value, allowed)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileExtension(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestMIMEType(t *testing.T) {
	t.Parallel()
	allowed := []string{"image/jpeg", "image/png", "application/pdf"}
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"jpeg", "image/jpeg", false},
		{"png", "image/png", false},
		{"pdf", "application/pdf", false},
		{"case insensitive", "IMAGE/JPEG", false},
		{"with spaces", "  image/jpeg  ", false},
		{"not allowed", "text/html", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := MIMEType("content", tt.value, allowed)
			if (err != nil) != tt.wantErr {
				t.Errorf("MIMEType(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func TestFileSize(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		size     int64
		maxBytes int64
		wantErr  bool
	}{
		{"within limit", 1024, 5000, false},
		{"exact limit", 5000, 5000, false},
		{"zero", 0, 5000, false},
		{"over limit", 5001, 5000, true},
		{"negative", -1, 5000, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := FileSize("file", tt.size, tt.maxBytes)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileSize(%d, %d) error = %v, wantErr %v", tt.size, tt.maxBytes, err, tt.wantErr)
			}
		})
	}
}

func TestSafeFilename(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		value   string
		wantErr bool
	}{
		{"simple", "photo.jpg", false},
		{"with dashes", "my-file-2024.txt", false},
		{"with spaces", "my file.txt", false},
		{"empty", "", true},
		{"dot", ".", true},
		{"dotdot", "..", true},
		{"forward slash", "path/file.txt", true},
		{"backslash", "path\\file.txt", true},
		{"traversal", "..hidden", true},
		{"control char", "file\x00.txt", true},
		{"tab", "file\t.txt", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := SafeFilename("file", tt.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SafeFilename(%q) error = %v, wantErr %v", tt.value, err, tt.wantErr)
			}
		})
	}
}

func BenchmarkFileExtension(b *testing.B) {
	allowed := []string{".jpg", ".png", ".gif", ".webp", ".svg"}
	for range b.N {
		_ = FileExtension("file", "photo.png", allowed)
	}
}

func BenchmarkSafeFilename(b *testing.B) {
	for range b.N {
		_ = SafeFilename("file", "my-photo-2024.jpg")
	}
}

func FuzzSafeFilename(f *testing.F) {
	f.Add("photo.jpg")
	f.Add("")
	f.Add("../etc/passwd")
	f.Add(".")
	f.Fuzz(func(t *testing.T, s string) {
		_ = SafeFilename("file", s)
	})
}
