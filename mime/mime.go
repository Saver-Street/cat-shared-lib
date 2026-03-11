package mime

import (
"mime"
"net/http"
"path/filepath"
"strings"
)

// Common MIME type constants.
const (
JSON           = "application/json"
XML            = "application/xml"
HTML           = "text/html"
Plain          = "text/plain"
CSS            = "text/css"
JavaScript     = "application/javascript"
PDF            = "application/pdf"
PNG            = "image/png"
JPEG           = "image/jpeg"
GIF            = "image/gif"
SVG            = "image/svg+xml"
WebP           = "image/webp"
OctetStream    = "application/octet-stream"
FormURLEncoded = "application/x-www-form-urlencoded"
MultipartForm  = "multipart/form-data"
CSV            = "text/csv"
Markdown       = "text/markdown"
)

// FromExtension returns the MIME type for the given file extension.
// The extension should include the dot (e.g., ".json").
// Returns "application/octet-stream" for unknown extensions.
func FromExtension(ext string) string {
if !strings.HasPrefix(ext, ".") {
ext = "." + ext
}
t := mime.TypeByExtension(ext)
if t == "" {
return OctetStream
}
// Remove parameters (e.g., charset)
if idx := strings.IndexByte(t, ';'); idx != -1 {
t = strings.TrimSpace(t[:idx])
}
return t
}

// FromFilename returns the MIME type for the given filename based on its extension.
func FromFilename(name string) string {
ext := filepath.Ext(name)
if ext == "" {
return OctetStream
}
return FromExtension(ext)
}

// FromBytes detects the MIME type of data by inspecting its content.
// Uses http.DetectContentType under the hood.
func FromBytes(data []byte) string {
return http.DetectContentType(data)
}

// IsText returns true if the MIME type is a text type.
func IsText(mimeType string) bool {
return strings.HasPrefix(baseType(mimeType), "text/")
}

// IsImage returns true if the MIME type is an image type.
func IsImage(mimeType string) bool {
return strings.HasPrefix(baseType(mimeType), "image/")
}

// IsAudio returns true if the MIME type is an audio type.
func IsAudio(mimeType string) bool {
return strings.HasPrefix(baseType(mimeType), "audio/")
}

// IsVideo returns true if the MIME type is a video type.
func IsVideo(mimeType string) bool {
return strings.HasPrefix(baseType(mimeType), "video/")
}

// IsJSON returns true if the MIME type is JSON.
func IsJSON(mimeType string) bool {
base := baseType(mimeType)
return base == JSON || strings.HasSuffix(base, "+json")
}

// IsXML returns true if the MIME type is XML.
func IsXML(mimeType string) bool {
base := baseType(mimeType)
return base == XML || base == "text/xml" || strings.HasSuffix(base, "+xml")
}

// Extension returns the preferred file extension for the given MIME type.
// Returns empty string for unknown types.
func Extension(mimeType string) string {
exts, err := mime.ExtensionsByType(mimeType)
if err != nil || len(exts) == 0 {
return ""
}
return exts[0]
}

// baseType strips parameters from a MIME type (e.g., "text/html; charset=utf-8" -> "text/html").
func baseType(mimeType string) string {
if idx := strings.IndexByte(mimeType, ';'); idx != -1 {
return strings.TrimSpace(mimeType[:idx])
}
return mimeType
}
