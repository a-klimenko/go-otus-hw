package main

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

var ErrDirectoryNotFound = errors.New("directory not found")

type Environment map[string]EnvValue

// EnvValue helps to distinguish between empty files and files with the first empty line.
type EnvValue struct {
	Value      string
	NeedRemove bool
}

// ReadDir reads a specified directory and returns map of env variables.
// Variables represented as files where filename is name of variable, file first line is a value.
func ReadDir(dir string) (Environment, error) {
	filesInfo, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, ErrDirectoryNotFound
	}
	env := make(Environment, len(filesInfo))

	for _, fileInfo := range filesInfo {
		var envValue EnvValue

		if fileInfo.Size() > 0 {
			file, err := os.Open(filepath.Join(dir, fileInfo.Name()))
			if err == nil {
				scanner := bufio.NewScanner(file)
				scanner.Scan()
				value := scanner.Bytes()
				value = bytes.ReplaceAll(value, []byte{0}, []byte("\n"))
				envValue.Value = strings.TrimRight(string(value), " \t")
			} else {
				os.Stderr.WriteString(err.Error())
			}
			if err := file.Close(); err != nil {
				os.Stderr.WriteString(err.Error())
			}
		} else {
			envValue.NeedRemove = true
		}

		env[fileInfo.Name()] = envValue
	}

	return env, nil
}
