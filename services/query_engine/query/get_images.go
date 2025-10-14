package query

import (
	"fmt"
	"query_engine/database"
	"sort"
	"utils"
)

func GetImages(db *database.DataBase, query string) ([]string, error) {
	words := utils.NormalizeQuery(query)

	// Map[imageURL] = {totalTF, matchedTerms}
	type imageScore struct {
		TotalTF      int
		MatchedTerms int
	}

	imageScores := make(map[string]*imageScore)

	for _, word := range words {
		imageIndex, err := db.GetImageTopX(word, 50)
		if err != nil {
			return nil, fmt.Errorf("error getting images for %s: %v", word, err)
		}

		for _, posting := range imageIndex.Postings {
			if _, exists := imageScores[posting.ImageUrl]; !exists {
				imageScores[posting.ImageUrl] = &imageScore{}
			}
			imageScores[posting.ImageUrl].TotalTF += posting.TermFrequency
			imageScores[posting.ImageUrl].MatchedTerms++
		}
	}

	// Convert to sortable slice
	type rankedImage struct {
		URL          string
		MatchedTerms int
		TotalTF      int
	}
	var ranked []rankedImage
	for url, score := range imageScores {
		ranked = append(ranked, rankedImage{
			URL:          url,
			MatchedTerms: score.MatchedTerms,
			TotalTF:      score.TotalTF,
		})
	}

	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].MatchedTerms == ranked[j].MatchedTerms {
			return ranked[i].TotalTF > ranked[j].TotalTF
		}
		return ranked[i].MatchedTerms > ranked[j].MatchedTerms
	})

	result := make([]string, 0, len(ranked))
	for _, r := range ranked {
		result = append(result, r.URL)
	}

	return result, nil
}
