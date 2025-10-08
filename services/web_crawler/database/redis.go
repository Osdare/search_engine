package database

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"strconv"
	"utils"
	"web_crawler/types"
	"github.com/redis/go-redis/v9"
)

type DataBase struct {
	client *redis.Client
	ctx    context.Context
}

const pageTag = "page"
const domainTag = "domain"

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

func (db *DataBase) PageExists(normPageUrl string) (bool, error) {
	res, err := db.client.Exists(db.ctx, pageTag+":"+normPageUrl).Result()
	if err != nil {
		return false, err
	}

	if res < 1 {
		return false, nil
	}

	return true, nil
}

func (db *DataBase) AddPage(page types.Page) error {
	h := sha256.Sum256([]byte(page.Content))
	checksum := hex.EncodeToString(h[:])

	res, err := db.client.SIsMember(db.ctx, "contenthashes", checksum).Result()
	if err != nil {
		return err
	}
	if res {
		log.Printf("content from page %s already exists", page.NormUrl)
		return nil
	}
	db.client.SAdd(db.ctx, "contenthashes", checksum)

	outLinksKey := "outlinks:" + utils.HashUrl(page.NormUrl)

	//add outlinks
	db.client.SAdd(db.ctx, outLinksKey, page.OutLinks)

	//index for outlinks for deterministic fetching
	if err = db.client.RPush(db.ctx, "outlinks:index", outLinksKey).Err(); err != nil {
		return fmt.Errorf("error when trying to push %v to outlinks:index %v", outLinksKey, err)
	}

	//add backlinks
	for _, backlink := range page.OutLinks {
		db.client.SAdd(db.ctx, "backlinks:"+utils.HashUrl(backlink), utils.HashUrl(page.NormUrl))
	}

	return nil
}

func (db *DataBase) AddDomain(domain types.Domain) error {

	if domain.CrawlDelay == 0 {
		domain.CrawlDelay = 1
	}
	crawlDelay := strconv.FormatInt(domain.CrawlDelay, 10)
	lastCrawled := strconv.FormatInt(domain.LastCrawled, 10)

	hashFields := []string{
		"crawldelay", crawlDelay,
		"lastcrawled", lastCrawled,
	}

	err := db.client.HSet(db.ctx, domainTag+":"+domain.Name, hashFields).Err()
	if err != nil {
		return fmt.Errorf("could not add domain %v to database %v", domain.Name, err)
	}

	if len(domain.Allowed) > 0 {
		err = db.client.SAdd(db.ctx, domainTag+":"+domain.Name+":"+"allowed", domain.Allowed).Err()
		if err != nil {
			return fmt.Errorf("could not add allowed to set of domain: %v %v", domain.Name, err)
		}
	}

	if len(domain.Disallowed) > 0 {
		err = db.client.SAdd(db.ctx, domainTag+":"+domain.Name+":"+"disallowed", domain.Disallowed).Err()
		if err != nil {
			return fmt.Errorf("could not add disallowed to set of domain: %v %v", domain.Disallowed, err)
		}
	}

	return nil
}

func (db *DataBase) GetDomain(domainName string) (types.Domain, error) {

	res, err := db.client.HGetAll(db.ctx, domainTag+":"+domainName).Result()
	if err != nil {
		return types.Domain{}, fmt.Errorf("could not get domain %v %v", domainName, err)
	}

	crawlDelay, err := strconv.ParseInt(res["crawldelay"], 10, 64)
	if err != nil {
		return types.Domain{}, fmt.Errorf("crawldelay could not be converted to int %v %v", res["crawldelay"], err)
	}

	lastCrawled, err := strconv.ParseInt(res["lastcrawled"], 10, 64)
	if err != nil {
		return types.Domain{}, fmt.Errorf("lastcrawled could not be converted to int %v %v", res["lastcrawled"], err)
	}

	allowed, err := db.client.SMembers(db.ctx, domainTag+":"+domainName+":"+"allowed").Result()
	if err != nil {
		return types.Domain{}, fmt.Errorf("could not retrieve allowed for domain: %v %v", domainName, err)
	}

	disallowed, err := db.client.SMembers(db.ctx, domainTag+":"+domainName+":"+"disallowed").Result()
	if err != nil {
		return types.Domain{}, fmt.Errorf("could not retrieve disallowed for domain: %v %v", domainName, err)
	}

	return types.Domain{
		Name:        domainName,
		CrawlDelay:  crawlDelay,
		LastCrawled: lastCrawled,
		Allowed:     allowed,
		Disallowed:  disallowed,
	}, nil
}

func (db *DataBase) UpdateDomainLastCrawled(domain string, lastCrawled int64) error {

	err := db.client.HSet(db.ctx, domainTag+":"+domain, "lastcrawled", lastCrawled).Err()
	if err != nil {
		return fmt.Errorf("could not update lastcrawled for domain: %v %v", domain, err)
	}

	return nil
}

func (db *DataBase) DomainExists(domainName string) (bool, error) {
	res, err := db.client.Exists(db.ctx, domainTag+":"+domainName).Result()
	if err != nil {
		return false, fmt.Errorf("error when checking if domain: %v exists %v", domainName, err)
	}
	return res > 0, nil
}

// Queue stuff
func (db *DataBase) PushUrl(normUrl string) error {

	exists, err := db.client.SIsMember(db.ctx, "urlset", normUrl).Result()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	if exists {
		//log.Printf("found %v results for url: %v", exists, normUrl)
		return nil
	}

	res, err := db.client.SAdd(db.ctx, "urlset", normUrl).Result()
	if err != nil {
		return fmt.Errorf("could not add url %v to set %v", normUrl, err)
	}
	if res == 0 {
		log.Printf("could not add: %v to urlset res: %v", normUrl, res)
		return nil
	}

	err = db.client.LPush(db.ctx, "urlqueue", normUrl).Err()
	if err != nil {
		return fmt.Errorf("could not push %v to urlqueue %v", normUrl, err)
	}

	return nil
}

func (db *DataBase) PopUrl() (string, error) {
	res, err := db.client.BRPop(db.ctx, 0, "urlqueue").Result()
	if err != nil {
		return "", fmt.Errorf("urlqueue could not be popped %v", err)
	}

	//Got rid of this because we don't want to crawl a website twice
	//err = db.client.SRem(db.ctx, "urlset", res[1]).Err()
	//if err != nil {
	//	return "", fmt.Errorf("could not remove %v from urlset %v", res[1], err)
	//}

	return res[1], nil
}

func (db *DataBase) RemoveUrlFromSet(normUrl string) error {
	err := db.client.SRem(db.ctx, "urlset", normUrl).Err()
	if err != nil {
		return fmt.Errorf("could not remove %v from urlset %v", normUrl, err)
	}
	return nil
}

func (db *DataBase) UrlQueueLength() (int64, error) {
	res, err := db.client.LLen(db.ctx, "urlqueue").Result()
	if err != nil {
		return 0, fmt.Errorf("could not retrieve length of urlqueue %v", err)
	}

	return res, nil
}

func (db *DataBase) AddIndex(index types.InvertedIndex) error {
	for term, posting := range index {
		err := db.client.ZAdd(db.ctx, "index:"+term, redis.Z{Member: posting.NormUrl, Score: float64(posting.TermFrequency)}).Err()
		if err != nil {
			return fmt.Errorf("could not add index to database %v", err)
		}
	}
	return nil
}

func (db *DataBase) AddDocument(document types.Document) error {
	err := db.client.HSet(db.ctx, "document:"+utils.HashUrl(document.NormUrl), "url", document.NormUrl, "title", document.Title, "length", document.Length).Err()
	if err != nil {
		return fmt.Errorf("could not add document for url: %v to database %v", document.NormUrl, err)
	}

	err = db.client.Incr(db.ctx, "domain:count").Err()
	if err != nil {
		return fmt.Errorf("failed to increment domain:count %v", err)
	}

	return nil
}

func (db *DataBase) AddImageIndex(index types.ImageIndex) error {
	for term, postings := range index {
		key := "imageindex:" + term

		members := make([]redis.Z, len(postings))
		for i, posting := range postings {
			members[i] = redis.Z{
				Member: posting.ImageUrl,
				Score:  float64(posting.TermFrequency),
			}
		}

		if err := db.client.ZAdd(db.ctx, key, members...).Err(); err != nil {
			return fmt.Errorf("could not add image index to db %v", err)
		}
	}
	return nil
}
