package utilities

import (
	"slices"
	"strings"
	"utils"
	"web_crawler/consts"
	"web_crawler/types"

	"github.com/reiver/go-porterstemmer"
)

func IndexImage(image types.Image) map[string]int {
	m := make(map[string]int)

	w := strings.SplitSeq(image.Text, " ")
	for word := range w {

		//normalizing and stemming
		word = strings.ToLower(word)
		word = strings.TrimSpace(word)
		word = utils.RemovePunctuation(word)

		stem := porterstemmer.StemWithoutLowerCasing([]rune(word))

		if len(stem) >= 2 &&
			len(stem) <= 32 &&
			!slices.Contains(consts.StopWords, word) &&
			utils.IsAlphanumeric(string(stem)) {
			m[string(stem)]++
		}
	}
	

	return m
}
