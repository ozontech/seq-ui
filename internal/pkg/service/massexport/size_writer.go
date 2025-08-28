package massexport

import "io"

type SizeWriter struct {
	io.Writer
	size int
}

func NewSizeWriter(w io.Writer) *SizeWriter {
	return &SizeWriter{Writer: w}
}

func (w *SizeWriter) Write(p []byte) (int, error) {
	n, err := w.Writer.Write(p)
	w.size += n
	return n, err
}

func (w *SizeWriter) Size() int {
	return w.size
}
