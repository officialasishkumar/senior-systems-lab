package server

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func TestFrameRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	want := []byte(`{"topic":"ops","payload":"ok"}`)
	if err := writeFrame(&buf, want); err != nil {
		t.Fatalf("writeFrame failed: %v", err)
	}
	got, err := readFrame(&buf)
	if err != nil {
		t.Fatalf("readFrame failed: %v", err)
	}
	if string(got) != string(want) {
		t.Fatalf("frame mismatch: got %s want %s", got, want)
	}
}

func FuzzReadFrame(f *testing.F) {
	f.Add(uint32(2), []byte("ok"))
	f.Add(uint32(0), []byte(""))
	f.Fuzz(func(t *testing.T, size uint32, payload []byte) {
		var buf bytes.Buffer
		_ = binary.Write(&buf, binary.BigEndian, size)
		buf.Write(payload)
		_, _ = readFrame(&buf)
	})
}
