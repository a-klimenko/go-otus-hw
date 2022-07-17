package main

import (
	"errors"
	"io"
	"os"
	"time"

	"github.com/cheggaaa/pb/v3"
)

const BufferSize = 1

var (
	ErrUnsupportedFile       = errors.New("unsupported file")
	ErrOffsetExceedsFileSize = errors.New("offset exceeds file size")
)

func Copy(fromPath, toPath string, offset, limit int64) error {
	source, err := os.Open(fromPath)
	if err != nil {
		return err
	}
	defer source.Close()

	fileInfo, fileInfoErr := source.Stat()
	if fileInfoErr != nil {
		return fileInfoErr
	}

	if !fileInfo.Mode().IsRegular() {
		return ErrUnsupportedFile
	}

	fileSize := fileInfo.Size()
	if fileSize < offset {
		return ErrOffsetExceedsFileSize
	}

	dest, destErr := os.Create(toPath)
	if destErr != nil {
		return destErr
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

	buf := make([]byte, BufferSize)
	var progress int64
	for progress < target {
		read, readErr := source.Read(buf)
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			return readErr
		}

		_, writeErr := dest.Write(buf)
		if writeErr != nil {
			return writeErr
		}

		progress += int64(read)
		bar.Increment()
		time.Sleep(time.Millisecond)
	}

	return nil
}
