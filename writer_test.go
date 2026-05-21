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
	"io/fs"
	"os"
	"testing"
	"time"
)

func TestGlobalHeaderWrite(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteGlobalHeader(); err != nil {
		t.Fatalf("WriteGlobalHeader: %v", err)
	}

	want := []byte("!<arch>\n")
	if !bytes.Equal(buf.Bytes(), want) {
		t.Errorf("global header: got %q, want %q", buf.Bytes(), want)
	}
}

func TestSimpleFile(t *testing.T) {
	body := "Hello world!\n"
	hdr := &Header{
		Name:    "hello.txt",
		ModTime: time.Unix(1361157466, 0),
		Size:    int64(len(body)),
		Mode:    fs.FileMode(0644),
		Uid:     501,
		Gid:     20,
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

	fixture, err := os.ReadFile("./fixtures/hello.a")
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if !bytes.Equal(fixture, buf.Bytes()) {
		t.Errorf("output mismatch:\n got  %q\n want %q", buf.Bytes(), fixture)
	}
}

func TestWriteTooLong(t *testing.T) {
	hdr := &Header{Size: 1}

	var buf bytes.Buffer
	w := NewWriter(&buf)
	if err := w.WriteHeader(hdr); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	_, err := w.Write([]byte("Hello world!\n"))
	if !errors.Is(err, ErrWriteTooLong) {
		t.Errorf("expected ErrWriteTooLong, got %v", err)
	}
}
