package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(data string) (string, error) {
	var prev rune
	var result strings.Builder

	if data == "" {
		return "", nil
	}

	for i, val := range data {
		count, _ := strconv.Atoi(string(val))
		if i == 0 && unicode.IsDigit(val) {
			return "", ErrInvalidString
		}

		if i > 0 {
			if unicode.IsDigit(val) {
				if unicode.IsDigit(prev) {
					return "", ErrInvalidString
				}
				result.WriteString(strings.Repeat(string(prev), count))
			} else if !unicode.IsDigit(prev) {
				result.WriteString(string(prev))
			}
		}
		prev = val
	}
	runes := []rune(data)
	lastSym := runes[len(runes)-1]

	if !unicode.IsDigit(lastSym) {
		result.WriteString(string(lastSym))
	}

	return result.String(), nil
}
