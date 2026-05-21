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

// Package ar implements access to ar archive files.
//
// An ar archive is a sequence of file entries, each preceded by a fixed-size
// header. The format is used by Unix linkers for static libraries (.a files)
// and by Debian package management (.deb files).
//
// This package is modelled after the standard library's [archive/tar] package.
//
// # Reading
//
//	f, err := os.Open("archive.a")
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer f.Close()
//
//	r, err := ar.NewReader(f)
//	if err != nil {
//		log.Fatal(err)
//	}
//	for hdr, err := range r.All() {
//		if err != nil {
//			log.Fatal(err)
//		}
//		fmt.Printf("%s (%d bytes)\n", hdr.Name, hdr.Size)
//		io.Copy(io.Discard, r) // consume entry data
//	}
//
// # Writing
//
//	var buf bytes.Buffer
//	w := ar.NewWriter(&buf)
//	if err := w.WriteGlobalHeader(); err != nil {
//		log.Fatal(err)
//	}
//	hdr := &ar.Header{Name: "hello.txt", Size: 13, Mode: 0644}
//	if err := w.WriteHeader(hdr); err != nil {
//		log.Fatal(err)
//	}
//	if _, err := io.WriteString(w, "Hello, world!\n"); err != nil {
//		log.Fatal(err)
//	}
package ar
