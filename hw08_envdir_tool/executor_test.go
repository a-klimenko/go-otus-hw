package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRunCmd(t *testing.T) {
	t.Run("base cases", func(t *testing.T) {
		os.Setenv("UNSET", "myvalue")
		os.Setenv("REPLACE", "myvalue")

		env := Environment{
			"BAR":     EnvValue{Value: "bar", NeedRemove: false},
			"EMPTY":   EnvValue{Value: "", NeedRemove: false},
			"UNSET":   EnvValue{Value: "", NeedRemove: true},
			"REPLACE": EnvValue{Value: "othervalue", NeedRemove: false},
		}

		// source file
		srcFile, err := os.CreateTemp("", "*_source_file")
		if err != nil {
			fmt.Println(err)
		}
		defer os.Remove(srcFile.Name())
		if _, err := srcFile.Write([]byte("content   \nnewline")); err != nil {
			srcFile.Close()
			fmt.Println(err)
		}

		// destination file
		destFile, errDest := os.CreateTemp("", "*_dest_file")
		if errDest != nil {
			fmt.Println(errDest)
		}
		destFile.Close()
		defer os.Remove(destFile.Name())

		cmd := []string{
			"cp",
			srcFile.Name(),
			destFile.Name(),
		}

		envValue, ok := os.LookupEnv("UNSET")
		require.True(t, ok)
		require.Equal(t, "myvalue", envValue)

		envValue, ok = os.LookupEnv("REPLACE")
		require.True(t, ok)
		require.Equal(t, "myvalue", envValue)

		returnCode := RunCmd(cmd, env)
		require.Equal(t, 0, returnCode)

		envValue, ok = os.LookupEnv("BAR")
		require.True(t, ok)
		require.Equal(t, "bar", envValue)

		envValue, ok = os.LookupEnv("EMPTY")
		require.True(t, ok)
		require.Equal(t, "", envValue)

		envValue, ok = os.LookupEnv("UNSET")
		require.False(t, ok)
		require.Equal(t, "", envValue)

		envValue, ok = os.LookupEnv("REPLACE")
		require.True(t, ok)
		require.Equal(t, "othervalue", envValue)
	})
}
