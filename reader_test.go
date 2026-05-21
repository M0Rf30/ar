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
	"io"
	"io/fs"
	"os"
	"testing"
	"time"
)

func openFixture(t *testing.T, name string) *os.File {
	t.Helper()
	f, err := os.Open("./fixtures/" + name)
	if err != nil {
		t.Fatalf("open fixture %s: %v", name, err)
	}
	t.Cleanup(func() {
		if err := f.Close(); err != nil {
			t.Errorf("close fixture %s: %v", name, err)
		}
	})
	return f
}

func newReader(t *testing.T, name string) *Reader {
	t.Helper()
	r, err := NewReader(openFixture(t, name))
	if err != nil {
		t.Fatalf("NewReader %s: %v", name, err)
	}
	return r
}

func TestReadHeader(t *testing.T) {
	reader := newReader(t, "hello.a")
	header, err := reader.Next()
	if err != nil {
		t.Fatalf("Next: %v", err)
	}

	tests := []struct {
		field     string
		got, want any
	}{
		{"Name", header.Name, "hello.txt"},
		{"ModTime", header.ModTime, time.Unix(1361157466, 0)},
		{"Uid", header.Uid, 501},
		{"Gid", header.Gid, 20},
		{"Mode", header.Mode, fs.FileMode(0644)},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got %v, want %v", tt.field, tt.got, tt.want)
		}
	}
}

func TestReadBody(t *testing.T) {
	reader := newReader(t, "hello.a")
	if _, err := reader.Next(); err != nil && err != io.EOF {
		t.Fatalf("Next: %v", err)
	}

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err != nil {
		t.Fatalf("Copy: %v", err)
	}

	want := []byte("Hello world!\n")
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("body: got %q, want %q", buf.Bytes(), want)
	}
}

func TestReadMulti(t *testing.T) {
	reader := newReader(t, "multi_archive.a")

	var buf bytes.Buffer
	for hdr, err := range reader.All() {
		if err != nil {
			t.Fatalf("All: %v", err)
		}
		_ = hdr
		if _, err := io.Copy(&buf, reader); err != nil {
			t.Fatalf("Copy: %v", err)
		}
	}

	want := []byte("Hello world!\nI love lamp.\n")
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("multi body: got %q, want %q", buf.Bytes(), want)
	}
}
