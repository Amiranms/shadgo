//go:build !change

package externalsort

// package main

import (
	"bufio"
	"io"
	"strings"
)

type StringHeap []string

func (h StringHeap) Len() int           { return len(h) }
func (h StringHeap) Less(i, j int) bool { return h[i] < h[j] }
func (h StringHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *StringHeap) Push(x any) {
	*h = append(*h, x.(string))
}

func (h *StringHeap) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

type HeapItem struct {
	line   string
	reader *LineReader
}

type MergeHeap []HeapItem

func (mh MergeHeap) Len() int           { return len(mh) }
func (mh MergeHeap) Less(i, j int) bool { return mh[i].line < mh[j].line }
func (mh MergeHeap) Swap(i, j int)      { mh[i], mh[j] = mh[j], mh[i] }

func (mh *MergeHeap) Push(x any) {
	*mh = append(*mh, x.(HeapItem))
}

func (mh *MergeHeap) Pop() any {

	old := *mh
	n := len(old)
	x := old[n-1]
	*mh = old[0 : n-1]
	return x
}

type LineReader interface {
	ReadLine() (string, error)
}

type LineWriter interface {
	Write(l string) error
}

type eReader struct {
	r *bufio.Reader
}

func (er eReader) ReadLine() (string, error) {
	line, err := er.r.ReadString('\n')
	if err != nil {
		if err != io.EOF || len(line) == 0 {
			return "", err
		}
		return strings.TrimRight(line, "\n"), nil

	}

	return strings.TrimRight(line, "\n"), nil
}

type eWriter struct {
	w *bufio.Writer
}

func (ew eWriter) WriteLine(l string) error {
	_, err := ew.w.WriteString(l + "\n")
	return err
}

func (ew eWriter) Flush() error {
	return ew.w.Flush()
}

func (w eWriter) Write(l string) error {
	err := w.WriteLine(l)
	if err != nil {
		return err
	}

	return w.Flush()
}

var _ LineReader = eReader{}

var _ LineWriter = eWriter{}

func NewReader(r io.Reader) LineReader {
	readerWrapper := bufio.NewReader(r)
	return eReader{r: readerWrapper}
}

func NewWriter(w io.Writer) LineWriter {
	bufw := bufio.NewWriter(w)
	return eWriter{w: bufw}
}
