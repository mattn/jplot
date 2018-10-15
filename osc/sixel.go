package osc

import (
	"bytes"
	"image/png"
	"os"
	"sync"
	"time"

	"github.com/mattn/go-sixel"

	"golang.org/x/crypto/ssh/terminal"
)

var sixelEnabled = false

func init() {
	if os.Getenv("TERM_PROGRAM") != "iTerm.app" {
		sixelEnabled = checkSixel()
	}
}

func checkSixel() bool {
	s, err := terminal.MakeRaw(1)
	if err != nil {
		return false
	}
	defer terminal.Restore(1, s)
	_, err = os.Stdout.Write([]byte("\x1b[c"))
	if err != nil {
		return false
	}
	defer fileSetReadDeadline(os.Stdout, time.Time{})

	var b [100]byte
	n, err := os.Stdout.Read(b[:])
	if err != nil {
		return false
	}
	if !bytes.HasPrefix(b[:n], []byte("\x1b[?63;")) {
		return false
	}
	for _, t := range bytes.Split(b[4:n], []byte(";")) {
		if len(t) == 1 && t[0] == '4' {
			return true
		}
	}
	return false
}

type sixelWriter struct {
	once sync.Once
	enc  *sixel.Encoder
	buf  *bytes.Buffer
}

func (w *sixelWriter) init() {
	w.buf = &bytes.Buffer{}
	w.enc = sixel.NewEncoder(os.Stdout)
}

// Write writes the PNG image data into the imageWriter buffer.
func (w *sixelWriter) Write(p []byte) (n int, err error) {
	w.once.Do(w.init)
	return w.buf.Write(p)
}

// Close flushes the image to the terminal and close the writer.
func (w *sixelWriter) Close() error {
	w.once.Do(w.init)
	img, err := png.Decode(w.buf)
	if err != nil {
		return err
	}
	return w.enc.Encode(img)
}