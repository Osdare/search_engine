package database

import (
	"context"
	"fmt"
	"indexer/types"
	"strconv"
	"strings"

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

func (db *DataBase) GetPageNodes() ([]types.PageNode, error) {
	keys, err := db.client.LRange(db.ctx, "outlinks:index", 1, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("error for fetching outlinks in range %d-%d %v", 0, -1, err)
	}

	pageNodes := make([]types.PageNode, len(keys))
	//fetch associated links
	for i, key := range keys {
		outLinks, err := db.client.SMembers(db.ctx, key).Result()
		if err != nil {
			return nil, fmt.Errorf("error fetching key %v reason %v", key, err)
		}

		urlHash := strings.Split(key, ":")[1]

		backlinksKey := "backlinks:" + urlHash

		backLinks, err := db.client.SMembers(db.ctx, backlinksKey).Result()
		if err != nil {
			return nil, fmt.Errorf("error fetching key %v reason %v", backlinksKey, err)
		}

		pageNodes[i] = types.PageNode{
			Hash:      urlHash,
			OutLinks:  outLinks,
			BackLinks: backLinks,
		}
	}

	return pageNodes, nil
}

func (db *DataBase) AddPageRanks(pageRanks map[string]float64) error {
	for hash, score := range pageRanks {

		err := db.client.Set(db.ctx, "pagerank:"+hash, score, 0).Err()
		if err != nil {
			return fmt.Errorf("could not add %v to db %v", hash, err)
		}
	}

	return nil
}