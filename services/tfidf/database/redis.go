package database

import (
	"context"
	"fmt"
	"log"
	"math"
	querytypes "query_engine/types"
	"strconv"
	"strings"
	"utils"

	"github.com/redis/go-redis/v9"
)

type DataBase struct {
	client *redis.Client
	ctx    context.Context
}

func (db *DataBase) Connect(addr string, database string, password string) error {
	dbId, err := strconv.Atoi(database)
	if err != nil {
		return err
	}

	db.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		DB:       dbId,
		Password: password,
	})

	db.ctx = context.Background()

	_, err = db.client.Ping(db.ctx).Result()
	if err != nil {
		return fmt.Errorf("couldn't connect do db %v %v", addr, err)
	}

	return nil
}

func (db *DataBase) GetDocLength(normUrl string) (int64, error) {
	key := "document:" + utils.HashUrl(normUrl)
	r, err := db.client.HGet(db.ctx, key, "length").Result()
	if err != nil {
		return 0, fmt.Errorf("could not get length of document %v from db %v", normUrl, err)
	}

	docLength, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse field %v %v", r, err)
	}

	return docLength, nil

}

func (db *DataBase) GetIndices() ([]querytypes.WordIndex, error) {

	keys, err := db.client.Keys(db.ctx, "index:*").Result()
	if err != nil {
		return nil, fmt.Errorf("could not get all keys for index %v", err)
	}

	indices := make([]querytypes.WordIndex, len(keys))
	for i, key := range keys {
		r, err := db.client.ZRevRangeWithScores(db.ctx, key, 0, -1).Result()
		word := strings.Split(key, ":")[1]
		if err != nil {
			return nil, fmt.Errorf("could not retrieve indices for word %v from db %v", word, err)
		}

		index := querytypes.WordIndex{
			Word: word,
		}
		for _, z := range r {
			normUrl, ok := z.Member.(string)
			if !ok {
				return nil, fmt.Errorf("expected string member but got %T", z.Member)
			}

			index.Postings = append(index.Postings, querytypes.Posting{
				NormUrl:       normUrl,
				TermFrequency: int(z.Score),
			})
		}

		indices[i] = index
	}

	return indices, nil
}

func (db *DataBase) GetDocsCount() (int64, error) {
	r, err := db.client.Get(db.ctx, "domain:count").Result()
	if err != nil {
		return 0, fmt.Errorf("could not get domain count from db %v", err)
	}

	count, err := strconv.ParseInt(r, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse int %v", err)
	}

	return count, nil
}

func (db *DataBase) StreamIndices(batchSize int) error {
	docsCount, err := db.GetDocsCount()
	if err != nil {
		return err
	}

	generateTfidf := func(indices []querytypes.WordIndex) error {
		pipe := db.client.Pipeline()

		docMagnitudeSums := make(map[string]float64)
		for _, index := range indices {
			idf := inverseDocumentFrequency(int(docsCount), len(index.Postings))

			pipe.ZAdd(db.ctx, "idf", redis.Z{
				Member: index.Word,
				Score:  idf,
			})

			for _, posting := range index.Postings {
				docLength, err := db.GetDocLength(posting.NormUrl)
				if err != nil {
					return fmt.Errorf("failed to get doc length for %s: %w", posting.NormUrl, err)
				}

				tfidf := Tfidf(
					posting.TermFrequency,
					int(docLength),
					int(docsCount),
					len(index.Postings),
				)

				pipe.ZAdd(db.ctx, "tfidf:"+index.Word, redis.Z{
					Member: posting.NormUrl,
					Score:  tfidf,
				})

				docMagnitudeSums[posting.NormUrl] += tfidf * tfidf
			}
		}

		for doc, sumSq := range docMagnitudeSums {
			magnitude := math.Sqrt(sumSq)
			pipe.ZAdd(db.ctx, "doc:magnitude", redis.Z{
				Member: doc,
				Score:  magnitude,
			})
		}

		if _, err := pipe.Exec(db.ctx); err != nil {
			return fmt.Errorf("failed to execute redis pipeline: %w", err)
		}

		return nil
	}

	var cursor uint64
	for {
		log.Printf("Processing indices cursor: %d\n", cursor)
		keys, nextCursor, err := db.client.Scan(db.ctx, cursor, "index:*", int64(batchSize)).Result()
		if err != nil {
			return fmt.Errorf("could not scan keys: %v", err)
		}

		var batch []querytypes.WordIndex
		for _, key := range keys {
			r, err := db.client.ZRevRangeWithScores(db.ctx, key, 0, -1).Result()
			if err != nil {
				return fmt.Errorf("could not get ZSET for key %s: %v", key, err)
			}

			word := strings.TrimPrefix(key, "index:")
			index := querytypes.WordIndex{Word: word}

			for _, z := range r {
				normUrl, ok := z.Member.(string)
				if !ok {
					return fmt.Errorf("expected string member but got %T", z.Member)
				}
				index.Postings = append(index.Postings, querytypes.Posting{
					NormUrl:       normUrl,
					TermFrequency: int(z.Score),
				})
			}
			batch = append(batch, index)
		}

		if err := generateTfidf(batch); err != nil {
			return err
		}

		if nextCursor == 0 {
			break
		}
		cursor = nextCursor
	}
	return nil
}

func relativeFrequency(termFrequency int, docLength int) float64 {
	return float64(termFrequency) / float64(docLength)
}

func inverseDocumentFrequency(totalDocs int, docsWithTerm int) float64 {
	N := float64(totalDocs)
	dft := float64(docsWithTerm)
	idf := math.Log((1+N)/(1+dft)) + 1
	return idf
}

func Tfidf(termFrequency int, docLength int, totalDocs int, docsWithTerm int) float64 {
	tf := relativeFrequency(termFrequency, docLength)
	idf := inverseDocumentFrequency(totalDocs, docsWithTerm)
	return tf * idf
}
