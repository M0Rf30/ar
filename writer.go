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
	"errors"
	"io"
	"io/fs"
	"strconv"
)

// ErrWriteTooLong is returned when more bytes are written than declared in the Header.
var ErrWriteTooLong = errors.New("ar: write too long")

// Writer provides sequential writing of an ar archive.
//
// Example:
//
//	w := ar.NewWriter(out)
//	w.WriteGlobalHeader()
//	hdr := &ar.Header{Name: "hello.txt", Size: 13, Mode: 0644}
//	w.WriteHeader(hdr)
//	io.Copy(w, data)
type Writer struct {
	w   io.Writer
	nb  int64 // unwritten bytes for the current entry
	pad bool  // whether a padding byte must be written when nb reaches 0
}

// NewWriter creates a new Writer writing to w.
func NewWriter(w io.Writer) *Writer { return &Writer{w: w} }

func formatNumeric(b []byte, x int64) {
	copy(b, padRight(strconv.FormatInt(x, 10), len(b)))
}

// formatOctal encodes mode in BSD ar format: "100<octal>" padded to len(b).
func formatOctal(b []byte, x fs.FileMode) {
	copy(b, padRight("100"+strconv.FormatInt(int64(x), 8), len(b)))
}

// WriteGlobalHeader writes the ar global header. Must be called once before any entries.
func (aw *Writer) WriteGlobalHeader() error {
	_, err := io.WriteString(aw.w, globalHeader)
	return err
}

// WriteHeader writes the file header and prepares the writer to accept the file's data.
func (aw *Writer) WriteHeader(hdr *Header) error {
	aw.nb = hdr.Size
	aw.pad = hdr.Size%2 == 1
	header := make([]byte, headerByteSize)
	s := slicer(header)

	copy(s.next(16), padRight(hdr.Name, 16))
	formatNumeric(s.next(12), hdr.ModTime.Unix())
	formatNumeric(s.next(6), int64(hdr.Uid))
	formatNumeric(s.next(6), int64(hdr.Gid))
	formatOctal(s.next(8), hdr.Mode)
	formatNumeric(s.next(10), hdr.Size)
	copy(s.next(2), padRight("`\n", 2))

	_, err := aw.w.Write(header)
	return err
}

// Write writes data for the current entry.
// Returns ErrWriteTooLong if more bytes are written than declared in the Header.
func (aw *Writer) Write(b []byte) (n int, err error) {
	if int64(len(b)) > aw.nb {
		b = b[0:aw.nb]
		err = ErrWriteTooLong
	}
	n, werr := aw.w.Write(b)
	aw.nb -= int64(n)
	if werr != nil {
		return n, werr
	}
	// emit alignment byte once all declared bytes have been written
	if aw.nb == 0 && aw.pad {
		aw.pad = false
		if _, perr := aw.w.Write([]byte{'\n'}); perr != nil && err == nil {
			err = perr
		}
	}
	return
}
