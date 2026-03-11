package mime_test

import (
"fmt"

"github.com/Saver-Street/cat-shared-lib/mime"
)

func ExampleFromFilename() {
fmt.Println(mime.FromFilename("report.pdf"))
// Output: application/pdf
}

func ExampleIsImage() {
fmt.Println(mime.IsImage("image/png"))
fmt.Println(mime.IsImage("text/plain"))
// Output:
// true
// false
}

func ExampleFromBytes() {
// PNG magic bytes
data := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
fmt.Println(mime.FromBytes(data))
// Output: image/png
}
