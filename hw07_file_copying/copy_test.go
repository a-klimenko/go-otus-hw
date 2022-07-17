package main

import (
	"errors"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	t.Run("unsupported file", func(t *testing.T) {
		err := Copy("/dev/urandom", "out.txt", 0, 0)
		require.True(t, errors.Is(err, ErrUnsupportedFile))
	})

	t.Run("offset exceeds file size", func(t *testing.T) {
		inputFile, err := os.CreateTemp("", "go_copy_test.*.txt")
		if err != nil {
			log.Fatal(err)
		}
		defer os.Remove(inputFile.Name())

		outFile, errOut := os.CreateTemp("", "go_copy_test.*.out.txt")
		if errOut != nil {
			log.Fatal(errOut)
		}
		defer os.Remove(outFile.Name())

		if _, err := inputFile.Write([]byte("content")); err != nil {
			inputFile.Close()
			log.Fatal(err)
		}
		if err := inputFile.Close(); err != nil {
			log.Fatal(err)
		}

		err = Copy(inputFile.Name(), outFile.Name(), 10000, 0)
		require.True(t, errors.Is(err, ErrOffsetExceedsFileSize))
	})
}
