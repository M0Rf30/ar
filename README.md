# ar

Go library for reading and writing [ar](https://en.wikipedia.org/wiki/Ar_(Unix)) archive files (Unix static libraries, Debian `.deb` packages).

Modelled after the standard library's [`archive/tar`](https://pkg.go.dev/archive/tar) package.

## Install

```
go get github.com/m0rf30/ar
```

## Usage

### Reading

```go
f, err := os.Open("archive.a")
if err != nil {
    log.Fatal(err)
}
defer f.Close()

r, err := ar.NewReader(f)
if err != nil {
    log.Fatal(err)
}
for hdr, err := range r.All() {
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%s (%d bytes)\n", hdr.Name, hdr.Size)
    io.Copy(io.Discard, r) // consume entry before calling Next
}
```

### Writing

```go
var buf bytes.Buffer
w := ar.NewWriter(&buf)
if err := w.WriteGlobalHeader(); err != nil {
    log.Fatal(err)
}
hdr := &ar.Header{
    Name:    "hello.txt",
    Size:    13,
    Mode:    0644,
    ModTime: time.Now(),
}
if err := w.WriteHeader(hdr); err != nil {
    log.Fatal(err)
}
if _, err := io.WriteString(w, "Hello, world!\n"); err != nil {
    log.Fatal(err)
}
```

## License

MIT — see [COPYING](COPYING).
