package mime

import (
"testing"
)

func TestFromExtension_JSON(t *testing.T) {
if got := FromExtension(".json"); got != JSON {
t.Errorf("FromExtension(.json) = %q, want %q", got, JSON)
}
}

func TestFromExtension_HTML(t *testing.T) {
got := FromExtension(".html")
if got != "text/html" {
t.Errorf("FromExtension(.html) = %q, want text/html", got)
}
}

func TestFromExtension_CSS(t *testing.T) {
got := FromExtension(".css")
if got != "text/css" {
t.Errorf("FromExtension(.css) = %q, want text/css", got)
}
}

func TestFromExtension_PNG(t *testing.T) {
if got := FromExtension(".png"); got != PNG {
t.Errorf("FromExtension(.png) = %q, want %q", got, PNG)
}
}

func TestFromExtension_Unknown(t *testing.T) {
if got := FromExtension(".xyzzyqwerty"); got != OctetStream {
t.Errorf("FromExtension(.xyzzyqwerty) = %q, want %q", got, OctetStream)
}
}

func TestFromExtension_WithoutDot(t *testing.T) {
if got := FromExtension("json"); got != JSON {
t.Errorf("FromExtension(json) = %q, want %q", got, JSON)
}
}

func TestFromExtension_PDF(t *testing.T) {
if got := FromExtension(".pdf"); got != PDF {
t.Errorf("FromExtension(.pdf) = %q, want %q", got, PDF)
}
}

func TestFromExtension_SVG(t *testing.T) {
if got := FromExtension(".svg"); got != SVG {
t.Errorf("FromExtension(.svg) = %q, want %q", got, SVG)
}
}

func TestFromFilename_Go(t *testing.T) {
// .go is not a registered MIME type on all systems; just check non-empty
got := FromFilename("main.go")
if got == "" {
t.Error("FromFilename(main.go) returned empty string")
}
}

func TestFromFilename_PDF(t *testing.T) {
if got := FromFilename("report.pdf"); got != PDF {
t.Errorf("FromFilename(report.pdf) = %q, want %q", got, PDF)
}
}

func TestFromFilename_DotJSON(t *testing.T) {
if got := FromFilename("data.json"); got != JSON {
t.Errorf("FromFilename(data.json) = %q, want %q", got, JSON)
}
}

func TestFromFilename_NoExtension(t *testing.T) {
if got := FromFilename("Makefile"); got != OctetStream {
t.Errorf("FromFilename(Makefile) = %q, want %q", got, OctetStream)
}
}

func TestFromFilename_Hidden(t *testing.T) {
// Hidden file with no real extension
got := FromFilename(".gitignore")
// .gitignore has no registered MIME; result may vary
if got == "" {
t.Error("FromFilename(.gitignore) returned empty string")
}
}

func TestFromBytes_JSON(t *testing.T) {
data := []byte(`{"key": "value"}`)
got := FromBytes(data)
// http.DetectContentType returns text/plain for JSON-like data
if got == "" {
t.Error("FromBytes(json) returned empty string")
}
}

func TestFromBytes_HTML(t *testing.T) {
data := []byte(`<!DOCTYPE html><html><body>Hello</body></html>`)
got := FromBytes(data)
if got != "text/html; charset=utf-8" {
t.Errorf("FromBytes(html) = %q, want text/html; charset=utf-8", got)
}
}

func TestFromBytes_PNG(t *testing.T) {
// PNG magic bytes
data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
got := FromBytes(data)
if got != "image/png" {
t.Errorf("FromBytes(png magic) = %q, want image/png", got)
}
}

func TestFromBytes_Unknown(t *testing.T) {
data := []byte{0x00, 0x01, 0x02, 0x03}
got := FromBytes(data)
if got != "application/octet-stream" {
t.Errorf("FromBytes(unknown) = %q, want application/octet-stream", got)
}
}

func TestIsText_Plain(t *testing.T) {
if !IsText(Plain) {
t.Error("IsText(text/plain) should be true")
}
}

func TestIsText_HTML(t *testing.T) {
if !IsText(HTML) {
t.Error("IsText(text/html) should be true")
}
}

func TestIsText_WithParams(t *testing.T) {
if !IsText("text/html; charset=utf-8") {
t.Error("IsText(text/html; charset=utf-8) should be true")
}
}

func TestIsText_JSON(t *testing.T) {
if IsText(JSON) {
t.Error("IsText(application/json) should be false")
}
}

func TestIsImage_PNG(t *testing.T) {
if !IsImage(PNG) {
t.Error("IsImage(image/png) should be true")
}
}

func TestIsImage_WithParams(t *testing.T) {
if !IsImage("image/jpeg; quality=80") {
t.Error("IsImage(image/jpeg; quality=80) should be true")
}
}

func TestIsImage_Text(t *testing.T) {
if IsImage(Plain) {
t.Error("IsImage(text/plain) should be false")
}
}

func TestIsAudio_True(t *testing.T) {
if !IsAudio("audio/mpeg") {
t.Error("IsAudio(audio/mpeg) should be true")
}
}

func TestIsAudio_False(t *testing.T) {
if IsAudio(JSON) {
t.Error("IsAudio(application/json) should be false")
}
}

func TestIsVideo_True(t *testing.T) {
if !IsVideo("video/mp4") {
t.Error("IsVideo(video/mp4) should be true")
}
}

func TestIsVideo_False(t *testing.T) {
if IsVideo(PNG) {
t.Error("IsVideo(image/png) should be false")
}
}

func TestIsJSON_ApplicationJSON(t *testing.T) {
if !IsJSON(JSON) {
t.Error("IsJSON(application/json) should be true")
}
}

func TestIsJSON_PlusJSON(t *testing.T) {
if !IsJSON("application/vnd.api+json") {
t.Error("IsJSON(application/vnd.api+json) should be true")
}
}

func TestIsJSON_NonJSON(t *testing.T) {
if IsJSON(Plain) {
t.Error("IsJSON(text/plain) should be false")
}
}

func TestIsXML_ApplicationXML(t *testing.T) {
if !IsXML(XML) {
t.Error("IsXML(application/xml) should be true")
}
}

func TestIsXML_TextXML(t *testing.T) {
if !IsXML("text/xml") {
t.Error("IsXML(text/xml) should be true")
}
}

func TestIsXML_PlusXML(t *testing.T) {
if !IsXML("application/atom+xml") {
t.Error("IsXML(application/atom+xml) should be true")
}
}

func TestIsXML_NonXML(t *testing.T) {
if IsXML(JSON) {
t.Error("IsXML(application/json) should be false")
}
}

func TestExtension_JSON(t *testing.T) {
ext := Extension(JSON)
if ext == "" {
t.Error("Extension(application/json) returned empty string")
}
}

func TestExtension_Unknown(t *testing.T) {
ext := Extension("application/x-nonexistent-type-zzzz")
if ext != "" {
t.Errorf("Extension(unknown) = %q, want empty string", ext)
}
}

func BenchmarkFromExtension(b *testing.B) {
for b.Loop() {
FromExtension(".json")
}
}

func BenchmarkFromBytes(b *testing.B) {
data := []byte(`{"key": "value", "number": 42}`)
for b.Loop() {
FromBytes(data)
}
}

func FuzzFromExtension(f *testing.F) {
f.Add(".json")
f.Add(".html")
f.Add(".xyz")
f.Add("")
f.Add("json")
f.Fuzz(func(t *testing.T, ext string) {
// Must not panic
result := FromExtension(ext)
if result == "" {
t.Error("FromExtension should never return empty string")
}
})
}
