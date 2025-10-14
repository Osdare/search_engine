package database

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"query_engine/types"
	"strconv"
	"time"
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

func (db *DataBase) GetTopXIndices(word string, x int64) (types.WordIndex, error) {
	r, err := db.client.ZRevRangeWithScores(db.ctx, "index:"+word, 0, x-1).Result()
	if err != nil {
		return types.WordIndex{}, fmt.Errorf("could not retrieve indices for word %v from db %v", word, err)
	}

	ten := types.WordIndex{Word: word}

	for _, z := range r {
		normUrl, ok := z.Member.(string)
		if !ok {
			return types.WordIndex{}, fmt.Errorf("expected string member but got %T", z.Member)
		}

		ten.Postings = append(ten.Postings, types.Posting{
			NormUrl:       normUrl,
			TermFrequency: int(z.Score),
		})
	}

	return ten, nil
}

func (db *DataBase) getIndex(prefix, word string) (types.Index, error) {
	r, err := db.client.ZRevRangeWithScores(db.ctx, prefix+word, 0, -1).Result()
	if err != nil {
		return types.Index{}, fmt.Errorf("could not retrieve indices for word %v from db: %v", word, err)
	}

	indices := types.Index{Word: word}

	for _, z := range r {
		normUrl, ok := z.Member.(string)
		if !ok {
			return types.Index{}, fmt.Errorf("expected string member but got %T", z.Member)
		}

		indices.Postings = append(indices.Postings, types.ScorePosting{
			NormUrl: normUrl,
			Score:   z.Score,
		})
	}

	return indices, nil
}

func (db *DataBase) GetTfidf(word string) (types.Index, error) {
	return db.getIndex("tfidf:", word)
}

func (db *DataBase) GetIdf(word string) (float64, error) {
	key := "idf"

	score, err := db.client.ZScore(db.ctx, key, word).Result()

	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("error fetching IDF for %s: %w", word, err)
	}

	return score, nil
}

func (db *DataBase) GetTfidfScoreForDoc(word string, link string) (float64, error) {
	key := fmt.Sprintf("tfidf:%s", word)
	score, err := db.client.ZScore(db.ctx, key, link).Result()
	if err == redis.Nil {
		return 0, nil
	} else if err != nil {
		return 0, fmt.Errorf("error fetching TF-IDF score for word %s in link %s: %w", word, link, err)
	}
	return score, nil
}

func (db *DataBase) GetDocumentMagnitude(link string) (float64, error) {
	key := "doc:magnitude"
	mag, err := db.client.ZScore(db.ctx, key, link).Result()
	if err == redis.Nil {
		return 0, fmt.Errorf("document magnitude not found for %s", link)
	} else if err != nil {
		return 0, fmt.Errorf("error fetching doc magnitude for %s: %w", link, err)
	}

	return mag, nil
}

func (db *DataBase) GetDocument(normUrl string) (types.Document, error) {
	r, err := db.client.HGetAll(db.ctx, "document:"+utils.HashUrl(normUrl)).Result()
	if err != nil {
		return types.Document{}, fmt.Errorf("could not get document %v from db %v", normUrl, err)
	}

	l := r["length"]
	length, err := strconv.Atoi(l)
	if err != nil {
		return types.Document{}, fmt.Errorf("could not parse %v %v", l, err)
	}

	d := types.Document{
		NormUrl: normUrl,
		Length:  length,
		Title:   r["title"],
	}

	return d, nil
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

func (db *DataBase) GetPageRank(url string) (float64, error) {
	key := "pagerank:" + utils.HashUrl(url)
	r, err := db.client.Get(db.ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("could not get pagerank for %v %v", key, err)
	}

	rank, err := strconv.ParseFloat(r, 64)
	if err != nil {
		return 0, fmt.Errorf("could not parse %v to float %v", r, err)
	}

	return rank, nil
}

func (db *DataBase) AddSearch(searchItem types.SearchItem) error {
	key := fmt.Sprintf("search:%s", utils.HashUrl(searchItem.NormalizedQuery))

	data, err := json.Marshal(searchItem)
	if err != nil {
		return fmt.Errorf("marshal searchItem failed %v", err)
	}

	ttl := 24 * time.Hour
	if err = db.client.Set(db.ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("could not push %v to db %v", key, err)
	}

	listKey := "search:recent"
	if err = db.client.LPush(db.ctx, listKey, key).Err(); err != nil {
		return fmt.Errorf("could not push %v to %v %v", key, listKey, err)
	}

	var maxSize int64 = 1000
	if err = db.client.LTrim(db.ctx, listKey, 0, maxSize-1).Err(); err != nil {
		return fmt.Errorf("could not trim %v %v", listKey, err)
	}

	length, err := db.client.LLen(db.ctx, listKey).Result()
	if err != nil {
		return fmt.Errorf("could not get length of %v %v", listKey, err)
	}

	if length > maxSize {
		oldKeys, _ := db.client.LRange(db.ctx, listKey, maxSize, -1).Result()
		if len(oldKeys) > 0 {
			db.client.LTrim(db.ctx, listKey, 0, maxSize-1)
			db.client.Del(db.ctx, oldKeys...)
		}
	}

	return nil
}

func (db *DataBase) GetSearch(normalizedQuery string) (types.SearchItem, error) {
	key := fmt.Sprintf("search:%s", utils.HashUrl(normalizedQuery))
	data, err := db.client.Get(db.ctx, key).Result()
	if err == redis.Nil {
		return types.SearchItem{}, nil
	}

	var searchItem types.SearchItem
	if err = json.Unmarshal([]byte(data), &searchItem); err != nil {
		return types.SearchItem{}, fmt.Errorf("could not unmarshal %v %v", data, err)
	}

	return searchItem, nil
}

func (db *DataBase) ComputeCosineSimilarity(words []string, linkScores map[string]float64) (map[string]float64, error) {

	// 1. Build Query Vector
	queryVec := make(map[string]float64)
	for _, word := range words {
		idf, err := db.GetIdf(word)
		if err != nil {
			return nil, err
		}
		tf := 1.0 // Simple TF for query
		queryVec[word] = tf * idf
	}

	// 2. Compute Query Vector Magnitude
	var queryMag float64
	for _, v := range queryVec {
		queryMag += v * v
	}
	queryMag = math.Sqrt(queryMag)

	// 3. Compute Cosine Similarity Per Document (Refactored)
	cosineScores := make(map[string]float64)
	if queryMag == 0 {
		return cosineScores, nil // No query magnitude means no score possible
	}

	for link := range linkScores {
		var dot float64

		// A. Dot Product: Iterate over query words and fetch the specific TF-IDF score for this link
		for word, qVal := range queryVec {
			// Fetch dVal (document TFIDF score for 'word' in 'link')
			dVal, err := db.GetTfidfScoreForDoc(word, link)
			if err != nil {
				// Log the error and continue to the next word
				fmt.Printf("Warning: Could not fetch TF-IDF score for word %s in link %s: %v\n", word, link, err)
				continue
			}
			dot += qVal * dVal
		}

		// B. Magnitude of Document Vector (OPTIMIZED LOOKUP)
		// This relies on the magnitude being pre-calculated during indexing.
		docMag, err := db.GetDocumentMagnitude(link)
		if err != nil {
			// If magnitude isn't found, we cannot calculate similarity.
			// Log this as a missing data issue.
			fmt.Printf("Warning: Could not fetch document magnitude for %s: %v\n", link, err)
			docMag = 0
		}

		// C. Cosine Similarity
		if docMag > 0 {
			cosineScores[link] = dot / (docMag * queryMag)
		} else {
			cosineScores[link] = 0
		}
	}

	return cosineScores, nil
}

func (db *DataBase) GetCandidateLinks(words []string, limit int64) (map[string]float64, error) {
	if len(words) == 0 {
		return make(map[string]float64), nil
	}

	// 1. Prepare keys for ZUNIONSTORE
	tfidfKeys := make([]string, len(words))
	for i, word := range words {
		tfidfKeys[i] = fmt.Sprintf("tfidf:%s", word)
	}

	// 2. Define a temporary key for the aggregated results
	tempKey := fmt.Sprintf("temp:union:%d", time.Now().UnixNano())

	// 3. Execute ZUNIONSTORE: sums the scores of the documents across all query words
	// The ZUNIONSTORE command is highly optimized for this task.
	err := db.client.ZUnionStore(db.ctx, tempKey, &redis.ZStore{
		Keys: tfidfKeys,
		// SUM aggregation means scores are summed: Score = TFIDF_w1 + TFIDF_w2 + ...
		Aggregate: "SUM",
	}).Err()

	if err != nil {
		return nil, fmt.Errorf("redis ZUNIONSTORE failed: %w", err)
	}

	// 4. Cleanup: Ensure the temporary key is deleted when done (best practice)
	defer db.client.Del(db.ctx, tempKey)

	// 5. Fetch the top candidates from the temporary result
	// We use the limit here to keep the candidate set small before the expensive cosine calculation.
	results, err := db.client.ZRevRangeWithScores(db.ctx, tempKey, 0, limit-1).Result()
	if err != nil {
		return nil, fmt.Errorf("redis ZRevRangeWithScores failed: %w", err)
	}

	// 6. Map results to the required linkScores format
	linkScores := make(map[string]float64)
	for _, z := range results {
		link, ok := z.Member.(string)
		if !ok {
			return nil, fmt.Errorf("expected string member in ZSET but got %T", z.Member)
		}
		linkScores[link] = z.Score // The score is the summed TF-IDF
	}

	return linkScores, nil
}