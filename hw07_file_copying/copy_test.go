package main

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCopy(t *testing.T) {
	t.Run("unsupported file", func(t *testing.T) {
		err := Copy("/dev/urandom", "out.txt", 0, 0)
		require.True(t, errors.Is(err, ErrUnsupportedFile))
	})

	t.Run("offset exceeds file size", func(t *testing.T) {
		err := Copy("testdata/input.txt", "out.txt", 10000, 0)
		require.True(t, errors.Is(err, ErrOffsetExceedsFileSize))
	})
}
