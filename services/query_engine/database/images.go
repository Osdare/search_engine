package database

import (
	"fmt"
	"query_engine/types"
)

func (db *DataBase) GetImageTopX(word string, x int64) (types.ImageIndex, error) {
	r, err := db.client.ZRevRangeWithScores(db.ctx, "imageindex:"+word, 0, x-1).Result()
	if err != nil {
		return types.ImageIndex{}, fmt.Errorf("could not get image %v from db %v", word, err)
	}

	results := types.ImageIndex{
		Word: word,
	}

	for _, z := range r {
		normUrl, ok := z.Member.(string)
		if !ok {
			return types.ImageIndex{}, fmt.Errorf("expected string member but got %T", z.Member)
		}

		results.Postings = append(results.Postings, types.ImagePosting{
			ImageUrl:      normUrl,
			TermFrequency: int(z.Score),
		})
	}

	return results, nil
}