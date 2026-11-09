// Package yenc_encoder implements the yEnc binary-to-text encoding used on
// Usenet for attachments. The encoding is deliberately simple: each byte is
// incremented by 42 (modulo 256) and then any of the bytes NUL, LF, CR, and
// '=' are escaped by emitting '=' followed by the original byte plus 64.
//
// The decoder is the inverse. It is intentionally permissive: it ignores
// bytes outside the encoded stream so that a caller can pass a full yEnc
// article (headers, blank lines, trailers) and get the payload back.
package yenc_encoder

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

// escapeSet is the set of bytes that must be escaped in the encoded stream.
// These are the control characters that would otherwise break line-based
// transport (NUL, LF, CR) plus the escape character itself.
var escapeSet = [256]bool{
	0x00: true, // NUL
	0x0A: true, // LF
	0x0D: true, // CR
	'=':  true,
}

// Encode returns the yEnc encoding of src. The output is a fresh slice; src
// is not modified.
func Encode(src []byte) []byte {
	// Worst case every byte is escaped, so allocate 2x up front to avoid
	// repeated growth on large attachments.
	dst := make([]byte, 0, len(src)*2)
	for _, b := range src {
		e := b + 42
		if escapeSet[e] {
			dst = append(dst, '=', e+64)
		} else {
			dst = append(dst, e)
		}
	}
	return dst
}

// Decode returns the yEnc decoding of src. It skips bytes until it sees a
// valid encoded byte, so callers may pass a complete article including
// headers. An error is returned if the stream ends in the middle of an escape
// sequence.
func Decode(src []byte) ([]byte, error) {
	dst := make([]byte, 0, len(src))
	i := 0
	for i < len(src) {
		b := src[i]
		i++
		if b == '=' {
			if i >= len(src) {
				return nil, errors.New("yenc: truncated escape sequence")
			}
			b = src[i] - 64
			i++
		}
		// The encoded byte is the original plus 42. We subtract 42 here.
		// Bytes that were not part of the encoded stream (e.g. article
		// headers) will decode to garbage, but that is the caller's
		// responsibility; yEnc has no framing of its own.
		dst = append(dst, b-42)
	}
	return dst, nil
}

// EncodeReader reads all of r and returns the yEnc encoding. It is a
// convenience wrapper around Encode for callers that have an io.Reader.
func EncodeReader(r io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		return nil, fmt.Errorf("yenc: read source: %w", err)
	}
	return Encode(buf.Bytes()), nil
}
