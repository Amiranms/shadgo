package otp

import (
	"io"
)

type CipherReader struct {
	r    io.Reader
	prng io.Reader
}

func (cr *CipherReader) Read(p []byte) (int, error) {
	n, err := cr.r.Read(p)
	if n > 0 {
		prngBytes := make([]byte, n)
		if _, errprng := io.ReadFull(cr.prng, prngBytes); errprng != nil {
			return 0, errprng
		}
		for i := 0; i < n; i++ {
			p[i] ^= prngBytes[i]
		}
	}
	return n, err
}

type CipherWriter struct {
	w    io.Writer
	prng io.Reader
}

func (cw *CipherWriter) Write(p []byte) (int, error) {
	prngBytes := make([]byte, len(p))
	if _, err := io.ReadFull(cw.prng, prngBytes); err != nil {
		return 0, err
	}
	ciphered := make([]byte, len(p))
	for i := 0; i < len(p); i++ {
		ciphered[i] = p[i] ^ prngBytes[i]
	}
	return cw.w.Write(ciphered)
}

func NewReader(r io.Reader, prng io.Reader) io.Reader {
	return &CipherReader{r: r, prng: prng}
}

func NewWriter(w io.Writer, prng io.Reader) io.Writer {
	return &CipherWriter{w: w, prng: prng}
}
