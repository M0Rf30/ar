/*
Copyright (c) 2013 Blake Smith <blakesmith0@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package ar

import (
	"bytes"
	"errors"
	"io"
	"io/fs"
	"testing"
	"time"
)

// TestNewReaderBadMagic verifies that NewReader rejects archives with an invalid magic header.
func TestNewReaderBadMagic(t *testing.T) {
	tests := []struct {
		name  string
		input []byte
	}{
		{"empty", []byte{}},
		{"truncated", []byte("!<arch>")},
		{"wrong magic", []byte("!<bad>\n\x00")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewReader(bytes.NewReader(tt.input))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			var fe *FormatError
			if !errors.As(err, &fe) {
				t.Errorf("expected *FormatError, got %T: %v", err, err)
			}
		})
	}
}

// TestRoundTrip writes a single entry and reads it back, verifying header and body.
func TestRoundTrip(t *testing.T) {
	body := "Hello, round-trip!\n" // 19 bytes (odd)
	hdr := &Header{
		Name:    "test.txt",
		ModTime: time.Unix(1700000000, 0),
		Uid:     1000,
		Gid:     1000,
		Mode:    fs.FileMode(0644),
		Size:    int64(len(body)),
	}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteGlobalHeader(); err != nil {
		t.Fatalf("WriteGlobalHeader: %v", err)
	}
	if err := w.WriteHeader(hdr); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	if _, err := w.Write([]byte(body)); err != nil {
		t.Fatalf("Write: %v", err)
	}

	r, err := NewReader(&buf)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	got, err := r.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}

	if got.Name != hdr.Name {
		t.Errorf("Name: got %q, want %q", got.Name, hdr.Name)
	}
	if got.ModTime != hdr.ModTime {
		t.Errorf("ModTime: got %v, want %v", got.ModTime, hdr.ModTime)
	}
	if got.Uid != hdr.Uid {
		t.Errorf("Uid: got %d, want %d", got.Uid, hdr.Uid)
	}
	if got.Gid != hdr.Gid {
		t.Errorf("Gid: got %d, want %d", got.Gid, hdr.Gid)
	}
	if got.Mode != hdr.Mode {
		t.Errorf("Mode: got %v, want %v", got.Mode, hdr.Mode)
	}
	if got.Size != hdr.Size {
		t.Errorf("Size: got %d, want %d", got.Size, hdr.Size)
	}

	var bodyBuf bytes.Buffer
	if _, err := io.Copy(&bodyBuf, r); err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if bodyBuf.String() != body {
		t.Errorf("body: got %q, want %q", bodyBuf.String(), body)
	}
}

// TestRoundTripEvenBody verifies correct behaviour with an even-sized entry (no padding byte).
func TestRoundTripEvenBody(t *testing.T) {
	body := "Hello!!\n" // 8 bytes (even)
	hdr := &Header{Name: "even.txt", Size: int64(len(body)), Mode: fs.FileMode(0644)}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteGlobalHeader(); err != nil {
		t.Fatalf("WriteGlobalHeader: %v", err)
	}
	if err := w.WriteHeader(hdr); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	if _, err := w.Write([]byte(body)); err != nil {
		t.Fatalf("Write: %v", err)
	}

	r, err := NewReader(&buf)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	if _, err := r.Next(); err != nil {
		t.Fatalf("Next: %v", err)
	}
	var bodyBuf bytes.Buffer
	if _, err := io.Copy(&bodyBuf, r); err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if bodyBuf.String() != body {
		t.Errorf("body: got %q, want %q", bodyBuf.String(), body)
	}
}

// TestRoundTripMulti writes two entries (odd then even) and reads both back.
func TestRoundTripMulti(t *testing.T) {
	entries := []struct {
		name, body string
	}{
		{"odd.txt", "Hello!\n"},  // 7 bytes (odd)  → padding byte written
		{"even.txt", "World!\n"}, // 7 bytes (odd)  → padding byte written
		{"last.txt", "Done.\n"},  // 6 bytes (even) → no padding
	}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteGlobalHeader(); err != nil {
		t.Fatalf("WriteGlobalHeader: %v", err)
	}
	for _, e := range entries {
		hdr := &Header{Name: e.name, Size: int64(len(e.body)), Mode: fs.FileMode(0644)}
		if err := w.WriteHeader(hdr); err != nil {
			t.Fatalf("WriteHeader %s: %v", e.name, err)
		}
		if _, err := w.Write([]byte(e.body)); err != nil {
			t.Fatalf("Write %s: %v", e.name, err)
		}
	}

	r, err := NewReader(&buf)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	i := 0
	for hdr, err := range r.All() {
		if err != nil {
			t.Fatalf("All[%d]: %v", i, err)
		}
		if hdr.Name != entries[i].name {
			t.Errorf("[%d] Name: got %q, want %q", i, hdr.Name, entries[i].name)
		}
		var bodyBuf bytes.Buffer
		if _, err := io.Copy(&bodyBuf, r); err != nil {
			t.Fatalf("Copy[%d]: %v", i, err)
		}
		if bodyBuf.String() != entries[i].body {
			t.Errorf("[%d] body: got %q, want %q", i, bodyBuf.String(), entries[i].body)
		}
		i++
	}
	if i != len(entries) {
		t.Errorf("read %d entries, want %d", i, len(entries))
	}
}

// TestAllEarlyBreak verifies that breaking out of All() does not panic or block.
func TestAllEarlyBreak(t *testing.T) {
	entries := []string{"a.txt", "b.txt", "c.txt"}
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteGlobalHeader(); err != nil {
		t.Fatalf("WriteGlobalHeader: %v", err)
	}
	for _, name := range entries {
		hdr := &Header{Name: name, Size: 4, Mode: fs.FileMode(0644)}
		if err := w.WriteHeader(hdr); err != nil {
			t.Fatalf("WriteHeader: %v", err)
		}
		if _, err := w.Write([]byte("data")); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}

	r, err := NewReader(&buf)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	seen := 0
	for hdr, err := range r.All() {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		_ = hdr
		seen++
		break // stop after first entry
	}
	if seen != 1 {
		t.Errorf("expected 1 entry before break, got %d", seen)
	}
}

// TestWriteMultiCall verifies that padding is emitted correctly when Write is called
// multiple times for a single entry (exercises the hdr.Size-based parity fix).
func TestWriteMultiCall(t *testing.T) {
	// 7-byte body (odd) split across two Write calls: 4 + 3
	body := "Hello!\n"
	hdr := &Header{Name: "split.txt", Size: int64(len(body)), Mode: fs.FileMode(0644)}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteGlobalHeader(); err != nil {
		t.Fatalf("WriteGlobalHeader: %v", err)
	}
	if err := w.WriteHeader(hdr); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	if _, err := w.Write([]byte("Hell")); err != nil {
		t.Fatalf("Write part 1: %v", err)
	}
	if _, err := w.Write([]byte("o!\n")); err != nil {
		t.Fatalf("Write part 2: %v", err)
	}

	r, err := NewReader(&buf)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	if _, err := r.Next(); err != nil {
		t.Fatalf("Next: %v", err)
	}
	var bodyBuf bytes.Buffer
	if _, err := io.Copy(&bodyBuf, r); err != nil {
		t.Fatalf("Copy: %v", err)
	}
	if bodyBuf.String() != body {
		t.Errorf("body: got %q, want %q", bodyBuf.String(), body)
	}
}
