package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	tmpDirPattern = "*_go_env_dir_reader_test"
	tmpFilPattern = "*_env_file_test"
)

func TestReadDir(t *testing.T) {
	t.Run("basic cases", func(t *testing.T) {
		expected := make(Environment, 3)
		envDir, err := os.MkdirTemp("", tmpDirPattern)
		if err != nil {
			fmt.Println(err)
		}
		defer os.RemoveAll(envDir) // clean up

		// regular variable
		envFile, err := os.CreateTemp(envDir, tmpFilPattern)
		if err != nil {
			fmt.Println(err)
		}
		defer os.Remove(envFile.Name())
		if _, err := envFile.Write([]byte("content   \nnewline")); err != nil {
			envFile.Close()
			fmt.Println(err)
		}
		expected[filepath.Base(envFile.Name())] = EnvValue{Value: "content", NeedRemove: false}

		// empty file case
		envEmptyFile, errEmptyFile := os.CreateTemp(envDir, tmpFilPattern)
		if errEmptyFile != nil {
			fmt.Println(errEmptyFile)
		}
		envEmptyFile.Close()
		defer os.Remove(envEmptyFile.Name())
		expected[filepath.Base(envEmptyFile.Name())] = EnvValue{Value: "", NeedRemove: true}

		// file with terminal zeros
		envTermFile, errTermFile := os.CreateTemp(envDir, tmpFilPattern)
		if errTermFile != nil {
			fmt.Println(errTermFile)
		}
		defer os.Remove(envTermFile.Name())
		termContent := append([]byte("content"), byte(0))
		if _, err := envTermFile.Write(termContent); err != nil {
			envTermFile.Close()
			fmt.Println(err)
		}
		expected[filepath.Base(envTermFile.Name())] = EnvValue{Value: "content\n", NeedRemove: false}

		result, errResult := ReadDir(envDir)
		require.NoError(t, errResult)
		require.Equal(t, expected, result)
	})

	t.Run("empty folder", func(t *testing.T) {
		envDir, err := os.MkdirTemp("", tmpDirPattern)
		if err != nil {
			fmt.Println(err)
		}
		defer os.RemoveAll(envDir) // clean up

		result, errResult := ReadDir(envDir)
		require.NoError(t, errResult)
		require.Equal(t, make(Environment, 0), result)
	})

	t.Run("non-existing folder", func(t *testing.T) {
		envDir, err := os.MkdirTemp("", tmpDirPattern)
		if err != nil {
			fmt.Println(err)
		}
		os.RemoveAll(envDir) // clean up

		result, errResult := ReadDir(envDir)
		require.True(t, errors.Is(errResult, ErrDirectoryNotFound))
		require.Nil(t, result)
	})
}
