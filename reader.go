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
	"io"
	"io/fs"
	"iter"
	"strconv"
	"time"
)

// Reader provides sequential read access to an ar archive.
//
// Example:
//
//	reader, err := ar.NewReader(f)
//	if err != nil {
//		return err
//	}
//	for hdr, err := range reader.All() {
//		if err != nil {
//			return err
//		}
//		var buf bytes.Buffer
//		io.Copy(&buf, reader)
//	}
type Reader struct {
	r   io.Reader
	nb  int64
	pad int64
}

// NewReader creates a new Reader reading from r.
// It consumes and validates the global ar header.
func NewReader(r io.Reader) (*Reader, error) {
	var magic [8]byte
	if _, err := io.ReadFull(r, magic[:]); err != nil {
		return nil, err
	}
	if string(magic[:]) != globalHeader {
		return nil, &FormatError{"invalid ar magic"}
	}
	return &Reader{r: r}, nil
}

// FormatError is returned when the archive format is invalid.
type FormatError struct {
	msg string
}

func (e *FormatError) Error() string { return "ar: " + e.msg }

func parseNumeric(b []byte) int64 {
	s := trimRight(b)
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// parseOctal parses the BSD ar octal mode field.
// The writer encodes mode as "100<octal>" (e.g. "100644"), so we strip the "100" prefix.
func parseOctal(b []byte) fs.FileMode {
	s := trimRight(b)
	if len(s) > 3 {
		s = s[3:]
	}
	n, _ := strconv.ParseInt(s, 8, 64)
	return fs.FileMode(n)
}

func (rd *Reader) skipUnread() error {
	skip := rd.nb + rd.pad
	rd.nb, rd.pad = 0, 0
	if skip == 0 {
		return nil
	}
	if seeker, ok := rd.r.(io.Seeker); ok {
		_, err := seeker.Seek(skip, io.SeekCurrent)
		return err
	}
	_, err := io.CopyN(io.Discard, rd.r, skip)
	return err
}

func (rd *Reader) readHeader() (*Header, error) {
	headerBuf := make([]byte, headerByteSize)
	if _, err := io.ReadFull(rd.r, headerBuf); err != nil {
		return nil, err
	}

	s := slicer(headerBuf)
	hdr := &Header{
		Name:    trimRight(s.next(16)),
		ModTime: time.Unix(parseNumeric(s.next(12)), 0),
		Uid:     int(parseNumeric(s.next(6))),
		Gid:     int(parseNumeric(s.next(6))),
		Mode:    parseOctal(s.next(8)),
		Size:    parseNumeric(s.next(10)),
	}

	rd.nb = hdr.Size
	if hdr.Size%2 == 1 {
		rd.pad = 1
	}

	return hdr, nil
}

// Next advances to the next file entry in the archive.
// Returns io.EOF when there are no more entries.
func (rd *Reader) Next() (*Header, error) {
	if err := rd.skipUnread(); err != nil {
		return nil, err
	}
	return rd.readHeader()
}

// All returns an iterator over all entries in the archive.
// The caller must read or discard each entry's data before the next iteration.
//
// Example:
//
//	for hdr, err := range reader.All() {
//		if err != nil {
//			return err
//		}
//		fmt.Println(hdr.Name)
//	}
func (rd *Reader) All() iter.Seq2[*Header, error] {
	return func(yield func(*Header, error) bool) {
		for {
			hdr, err := rd.Next()
			if err == io.EOF {
				return
			}
			if !yield(hdr, err) || err != nil {
				return
			}
		}
	}
}

// Read reads from the current entry in the archive.
func (rd *Reader) Read(b []byte) (n int, err error) {
	if rd.nb == 0 {
		return 0, io.EOF
	}
	if int64(len(b)) > rd.nb {
		b = b[0:rd.nb]
	}
	n, err = rd.r.Read(b)
	rd.nb -= int64(n)
	return
}
