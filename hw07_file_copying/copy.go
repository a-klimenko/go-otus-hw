package main

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/cheggaaa/pb/v3"
)

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	var source *os.File
	var dest *os.File
	var err error

	source, err = os.Open(fromPath)
	if err != nil {
		return err
	}
	defer source.Close()
	var fileInfo os.FileInfo
	fileInfo, err = source.Stat()
	if err != nil {
		return err
	}

	if !fileInfo.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	fileSize := fileInfo.Size()
	if fileSize < offset {
		return ErrOffsetExceedsFileSize
	}

	dest, err = os.Create(toPath)
	if err != nil {
		return err
	}
	defer dest.Close()

	if _, err := source.Seek(offset, io.SeekStart); err != nil {
		return err
	}

	target := fileSize - offset
	if target > limit && limit > 0 {
		target = limit
	}

	bar := pb.StartNew(int(target))
	defer bar.Finish()
	bar.SetWriter(os.Stdout)
	bar.Set(pb.Bytes, true)
	bar.Set(pb.SIBytesPrefix, true)

	buf := make([]byte, 1)
	var progress int64
	for progress < target {
		read, readErr := source.Read(buf)
		progress += int64(read)

		bar.Increment()
		time.Sleep(time.Millisecond)

		_, writeErr := dest.Write(buf)
		if writeErr != nil {
			return writeErr
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}
	}

	return nil
}
