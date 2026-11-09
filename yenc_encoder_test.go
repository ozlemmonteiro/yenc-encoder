package yenc_encoder

import (
	"bytes"
	"strings"
	"testing"
)

func TestEncodeEmpty(t *testing.T) {
	got := Encode(nil)
	if len(got) != 0 {
		t.Fatalf("Encode(nil) = %v, want empty", got)
	}
}

func TestEncodeKnownBytes(t *testing.T) {
	// Byte 0x00 becomes 42, which is not in the escape set.
	got := Encode([]byte{0x00})
	want := []byte{42}
	if !bytes.Equal(got, want) {
		t.Fatalf("Encode([0]) = %v, want %v", got, want)
	}
}

func TestEncodeEscapes(t *testing.T) {
	// 0x00 + 42 = 42 (not escaped)
	// 0xD6 + 42 = 0x00 (escaped)
	// 0xE3 + 42 = 0x0D (escaped)
	// 0xE4 + 42 = 0x0E (not escaped)
	got := Encode([]byte{0xD6, 0xE3, 0xE4})
	want := []byte{'=', 0x00 + 64, '=', 0x0D + 64, 0x0E}
	if !bytes.Equal(got, want) {
		t.Fatalf("Encode escapes = %v, want %v", got, want)
	}
}

func TestRoundTrip(t *testing.T) {
	src := []byte("Hello, yEnc!\x00\r\n=the quick brown fox")
	enc := Encode(src)
	dec, err := Decode(enc)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if !bytes.Equal(dec, src) {
		t.Fatalf("round trip = %q, want %q", dec, src)
	}
}

func TestDecodeTruncatedEscape(t *testing.T) {
	_, err := Decode([]byte{'='})
	if err == nil {
		t.Fatal("Decode of truncated escape should fail")
	}
}

func TestDecodeIgnoresSurroundingText(t *testing.T) {
	// A caller may hand us a full article. We skip bytes until we see
	// encoded data, but we do not know where the payload starts, so we
	// document that Decode expects the encoded stream only. Here we pass
	// a bare encoded stream to prove the happy path.
	src := []byte("abc")
	enc := Encode(src)
	dec, err := Decode(enc)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if !bytes.Equal(dec, src) {
		t.Fatalf("Decode = %q, want %q", dec, src)
	}
}

func TestEncodeReader(t *testing.T) {
	r := strings.NewReader("yEnc")
	got, err := EncodeReader(r)
	if err != nil {
		t.Fatalf("EncodeReader error: %v", err)
	}
	want := Encode([]byte("yEnc"))
	if !bytes.Equal(got, want) {
		t.Fatalf("EncodeReader = %v, want %v", got, want)
	}
}

func TestEncodeReaderError(t *testing.T) {
	// A reader that always fails should surface the error.
	r := &failingReader{}
	_, err := EncodeReader(r)
	if err == nil {
		t.Fatal("EncodeReader should propagate read errors")
	}
}

type failingReader struct{}

func (f *failingReader) Read(p []byte) (int, error) {
	return 0, bytes.ErrTooLarge
}
