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
	"io/fs"
	"strings"
	"time"
)

const (
	headerByteSize = 60
	globalHeader   = "!<arch>\n"
)

// Header represents the metadata for a single file entry in an ar archive.
type Header struct {
	Name    string
	ModTime time.Time
	Uid     int
	Gid     int
	Mode    fs.FileMode
	Size    int64
}

type slicer []byte

func (sp *slicer) next(n int) []byte {
	s := *sp
	b := s[0:n]
	*sp = s[n:]
	return b
}

// trimRight strips trailing ASCII spaces from b and returns the result as a string.
func trimRight(b []byte) string {
	i := len(b) - 1
	for i > 0 && b[i] == ' ' {
		i--
	}
	return string(b[0 : i+1])
}

// padRight returns s padded with spaces on the right to exactly n bytes.
func padRight(s string, n int) string {
	if len(s) < n {
		s += strings.Repeat(" ", n-len(s))
	}
	return s
}
