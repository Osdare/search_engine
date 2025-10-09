package query

import (
	"query_engine/database"
	"sort"
	"utils"
)

type QueryConfig struct {
	IndexCount     int64
	UrlReturnCount int
}

func GetRelevantUrls(query string, db *database.DataBase, qf QueryConfig) ([]string, error) {
	words := utils.NormalizeQuery(query)

	candidateLinks, err := db.GetCandidateLinks(words, 100)
	if err != nil {
		return nil, err
	}

	cosineScores, err := db.ComputeCosineSimilarity(words, candidateLinks)
	if err != nil {
		return nil, err
	}

	type kv struct {
		key string
		val float64
	}

	results := make([]kv, 0)
	for url, score := range cosineScores {
		results = append(results, kv{
			key: url,
			val: score,
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].val > results[j].val
	})

	linksLength := min(len(results), qf.UrlReturnCount)

	links := make([]string, 0)
	for i := range linksLength {
		links = append(links, results[i].key)
	}

	return links, nil
}
