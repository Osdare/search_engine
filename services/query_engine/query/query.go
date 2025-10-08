package query

import (
	"fmt"
	"math"
	"query_engine/database"
	"query_engine/tfidf"
	"query_engine/types"
	"sort"
	"utils"
)

type QueryConfig struct {
	IndexCount     int64
	UrlReturnCount int
}


func GetUrlsFromQuery(query string, db *database.DataBase, qf QueryConfig) ([]string, error) {
	words := utils.NormalizeQuery(query)

	docsCount, err := db.GetDocsCount()
	if err != nil {
		return nil, err
	}

	//get indices for the words
	linkScores := make(map[string]float64)
	for _, word := range words {
		index, err := db.GetTopXIndices(word, qf.IndexCount)
		if err != nil {
			return nil, err
		}

		//iterate through postings
		for _, posting := range index.Postings {
			docLength, err := db.GetDocLength(posting.NormUrl)
			if err != nil {
				return nil, err
			}

			score := tfidf.Tfidf(posting.TermFrequency, int(docLength), int(docsCount), len(index.Postings))

			linkScores[posting.NormUrl] += score

		}
	}

	searchResults := make([]types.SearchResult, 0)
	tfidfScores := make(map[string]float64)
	pageRankScores := make(map[string]float64)

	for link, tfidf := range linkScores {
		pageRank, err := db.GetPageRank(link)
		if err != nil {
			continue
		}

		searchResults = append(searchResults, types.SearchResult{
			Url: link,
		})
		tfidfScores[link] = tfidf
		pageRankScores[link] = pageRank
	}

	normTFIDF := normalizeScores(tfidfScores)
	normPR := normalizeScores(pageRankScores)

	for i := range searchResults {
		link := searchResults[i].Url
		searchResults[i].FinalScore = 0.7*normTFIDF[link] + 0.3*normPR[link]
		fmt.Println("t", normTFIDF[link])
		fmt.Println("p", normPR[link])
	}

	sort.Slice(searchResults, func(i, j int) bool {
		return searchResults[i].FinalScore > searchResults[j].FinalScore
	})

	resultLinks := make([]string, qf.UrlReturnCount)
	for i := range qf.UrlReturnCount {
		resultLinks[i] = searchResults[i].Url
	}

	return resultLinks, nil
}

func normalizeScores(scores map[string]float64) map[string]float64 {
	if len(scores) == 0 {
		return scores
	}

	var min, max float64 = math.MaxFloat64, -math.MaxFloat64
	for _, score := range scores {
		if score < min {
			min = score
		}
		if score > max {
			max = score
		}
	}

	normalized := make(map[string]float64)
	rangeDiff := max - min
	if rangeDiff == 0 {
		for hash := range scores {
			normalized[hash] = 0.5
		}
		return normalized
	}

	for hash, score := range scores {
		normalized[hash] = (score - min) / rangeDiff
	}
	return normalized
}
