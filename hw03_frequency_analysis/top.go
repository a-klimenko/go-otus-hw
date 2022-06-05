package hw03frequencyanalysis

import (
	"sort"
	"strings"
)

type Pair struct {
	Word  string
	Count int
}

func Top10(data string) []string {
	if data == "" {
		return []string{}
	}

	wordFrequencies := make(map[string]int)
	words := strings.Fields(data)

	for _, word := range words {
		wordFrequencies[word]++
	}

	pl := make([]Pair, 0, len(wordFrequencies))

	for word, count := range wordFrequencies {
		pl = append(pl, Pair{word, count})
	}

	sort.Slice(pl, func(i, j int) bool {
		switch {
		case pl[i].Count != pl[j].Count:
			return pl[i].Count > pl[j].Count
		default:
			return pl[i].Word < pl[j].Word
		}
	})

	result := make([]string, 0, 10)
	for idx, pair := range pl {
		result = append(result, pair.Word)
		if idx == 9 {
			break
		}
	}

	return result
}
