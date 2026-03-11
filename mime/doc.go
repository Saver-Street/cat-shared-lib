// Package mime provides utilities for MIME type detection and classification.
//
// It supports detection from file extensions, filenames, and raw byte content,
// as well as classification helpers for common media categories (text, image,
// audio, video) and structured formats (JSON, XML).
//
//mimeType := mime.FromFilename("report.pdf")  // "application/pdf"
//isImg    := mime.IsImage("image/png")         // true
//detected := mime.FromBytes(rawData)           // e.g. "image/png"
package mime
