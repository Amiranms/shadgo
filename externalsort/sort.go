//go:build !solution

package externalsort

import (
	"container/heap"
	"io"
	"os"
)

func SortFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	reader := NewReader(file)
	sortedStr := &StringHeap{}
	heap.Init(sortedStr)
	for {
		line, err := reader.ReadLine()
		if err != nil {
			break
		}
		heap.Push(sortedStr, line)
	}
	file, err = os.Create(path)
	defer file.Close()

	writer := NewWriter(file)
	for len(*sortedStr) != 0 {
		element := heap.Pop(sortedStr).(string)
		writer.Write(element)
	}
	return nil
}

func Merge(w LineWriter, readers ...LineReader) error {
	lines := &MergeHeap{}
	heap.Init(lines)
	for _, reader := range readers {
		line, err := reader.ReadLine()
		if err == io.EOF {
			continue
		}
		if err != nil {
			return err
		}
		hi := HeapItem{line: line, reader: &reader}
		heap.Push(lines, hi)
	}

	for lines.Len() > 0 {
		item := heap.Pop(lines).(HeapItem)
		err := w.Write(item.line)
		if err != nil {
			return err
		}
		nextReader := *item.reader
		nextline, readErr := nextReader.ReadLine()
		if readErr == io.EOF {
			continue
		}
		if readErr != nil {
			return readErr
		}
		heap.Push(lines, HeapItem{line: nextline, reader: item.reader})
	}
	return nil
}

func Sort(w io.Writer, in ...string) error {
	// по хорошему здесь было бы открыть их и не закрывать, но не страшно
	for _, txt := range in {
		SortFile(txt)
	}
	readers := make([]LineReader, 0, len(in))

	var openFiles []*os.File
	defer func() {
		for _, f := range openFiles {
			f.Close()
		}
	}()

	for _, path := range in {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		openFiles = append(openFiles, f)
		readers = append(readers, NewReader(f))
	}
	Merge(NewWriter(w), readers...)
	return nil
}
