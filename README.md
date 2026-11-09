# yenc-encoder

A small Go library that encodes and decodes the yEnc binary-to-text encoding used on Usenet for attachments.

## Usage

```go
package main

import (
	"fmt"

	yenc "yenc-encoder"
)

func main() {
	src := []byte("Hello, yEnc!")
	encoded := yenc.Encode(src)
	fmt.Printf("encoded: %q\n", encoded)

	decoded, err := yenc.Decode(encoded)
	if err != nil {
		panic(err)
	}
	fmt.Printf("decoded: %q\n", decoded)
}
```

## Why it exists

yEnc is a compact binary-to-text encoding that predates the modern web. It is still used by Usenet clients, and a dependency-free implementation is useful when you need to embed it in a larger tool without pulling in a full Usenet stack. The trade-off here is simplicity: the decoder does not attempt to locate the payload inside a full article, because yEnc has no framing of its own. It expects the encoded stream, or a stream that begins at the encoded data. If you have a full article, you must split it yourself before calling `Decode`.

## Edge case

The decoder treats any `=` in the input as the start of an escape sequence. A stray `=` that is not followed by a byte will return an error. This is intentional: a well-formed yEnc stream never contains a bare `=`.
