package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(data string) (string, error) {
	var prev string
	var result strings.Builder

	if data == "" {
		return "", nil
	}

	for i, val := range data {
		count, err := strconv.Atoi(string(val))
		if i == 0 && err == nil {
			return "", ErrInvalidString
		}

		if i > 0 {
			_, errPrev := strconv.Atoi(prev)
			if err == nil && errPrev == nil {
				return "", ErrInvalidString
			}

			if err == nil {
				result.WriteString(strings.Repeat(prev, count))
			} else if errPrev != nil {
				result.WriteString(prev)
			}
		}
		prev = string(val)
	}
	runes := []rune(data)
	lastSym := string(runes[len(runes)-1:])

	if _, err := strconv.Atoi(lastSym); err != nil {
		result.WriteString(lastSym)
	}

	return result.String(), nil
}
