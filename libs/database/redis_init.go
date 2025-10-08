package database

import (
	"context"
	"fmt"
	"strconv"

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