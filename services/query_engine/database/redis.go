package database

import (
	"context"
	"encoding/json"
	"fmt"
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

func (db *DataBase) GetImageTopX(word string, x int64) (types.ImageIndex, error) {
	r, err := db.client.ZRevRangeWithScores(db.ctx, "imageindex:"+word, 0, x-1).Result()
	if err != nil {
		return types.ImageIndex{}, fmt.Errorf("could not get image %v from db %v", word, err)
	}

	ten := types.ImageIndex{
		Word: word,
	}

	for _, z := range r {
		normUrl, ok := z.Member.(string)
		if !ok {
			return types.ImageIndex{}, fmt.Errorf("expected string member but got %T", z.Member)
		}

		ten.Postings = append(ten.Postings, types.ImagePosting{
			ImageUrl:      normUrl,
			TermFrequency: int(z.Score),
		})
	}

	return ten, nil
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
