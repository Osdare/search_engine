package crawler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	"utils"
	"web_crawler/database"
	"web_crawler/handlers"
	"web_crawler/parser"
	"web_crawler/types"
	"web_crawler/utilities"

	"golang.org/x/net/html"
)

// returns html as node
func Crawl(normUrl string) (*html.Node, string, error) {
	userAgent := os.Getenv("USER_AGENT")

	req, err := http.NewRequest("GET", normUrl, nil)
	if err != nil {
		return nil, "", fmt.Errorf("error from http request to %v %v", normUrl, err)
	}

	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("could not get url: %v %v", normUrl, err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("could not read body %v", err)
	}

	html, err := html.Parse(bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, "", fmt.Errorf("could not parse response body %v", err)
	}

	content := string(bodyBytes)

	return html, content, nil
}

func CrawlJob(db *database.DataBase) {

	start := time.Now()

	//get url from database
	link, err := db.PopUrl()
	if err != nil {
		log.Printf("could not pop url from db %v\n", err)
		return
	}

	u, err := url.Parse(link)
	if err != nil {
		log.Printf("could not parse url: %v %v\n", link, err)
		return
	}

	//check if domain exists and create a new one if not
	domainExists, err := db.DomainExists(u.Host)
	if err != nil {
		log.Printf("couldn not check if domain: %v exists %v\n", u.Host, err)
		return
	}

	var domain types.Domain
	if domainExists {
		domain, err = db.GetDomain(u.Host)
		if err != nil {
			log.Printf("could not get domain: %v %v\n", u.Host, err)
			return
		}
	} else {
		domain, err = handlers.GetRobotsFromDomain(u.Scheme + "://" + u.Host)
		if err != nil {
			err = db.AddDomain(domain)
			if err != nil {
				log.Printf("bruh moment %v\n", err)
			}
			log.Printf("could not get new domain: %v reason: %v", u.Scheme+"://"+u.Host, err)
			return
		}

		err = db.AddDomain(domain)
		if err != nil {
			log.Printf("could not add domain: %v to db. Reason: %v", u.Host, err)
			return
		}
	}

	//check domain and if we are allowed to crawl else return
	//and put url back in the queue
	canCrawl, reason, err := handlers.CanCrawl(link, domain)
	if err != nil {
		log.Printf("error from CanCrawl function %v\n", err)
		return
	}
	if !canCrawl {
		if reason == handlers.ReasonCrawlDelay {
			//remove link from set and put back in queue
			err = db.RemoveUrlFromSet(link)
			if err != nil {
				log.Printf("%v\n", err)
			}
			err = db.PushUrl(link)
			if err != nil {
				log.Printf("could not push url to queue. Reason: %v\n", err)
			}
		}
		return
	}

	err = db.UpdateDomainLastCrawled(domain.Name, utils.GetTimeInt())
	if err != nil {
		log.Println(err)
	}

	log.Printf("Crawling: %v", link)

	//crawl it
	html, content, err := Crawl(link)
	if err != nil {
		log.Printf("Could not crawl url: %v %v\n", link, err)
		return
	}

	//normalize urls and put new urls in database
	title, rawUrls, images, wordMap := parser.ParseBody(link, html)
	newUrls, err := utilities.NormalizeUrlSlice(link, rawUrls)
	if err != nil {
		log.Println(err)
		return
	}

	page := types.Page{
		NormUrl:  link,
		Content:  content,
		OutLinks: newUrls,
	}
	err = db.AddPage(page)
	if err != nil {
		log.Printf("page: %v could not be added to database %v\n", link, err)
	}

	for _, newUrl := range newUrls {
		//log.Printf("attempting to add url: %v to queue\n", newUrl)
		err = db.PushUrl(newUrl)
		if err != nil {
			log.Printf("Could not add url: %v to queue %v\n", newUrl, err)
		}
	}

	//add image indices
	imageIndex := types.ImageIndex{}
	for _, image := range images {
		m := utilities.IndexImage(image)

		for term, frequency := range m {
			imageUrl, err := utilities.NormalizeLink(link, image.ImageUrl)
			if err != nil {
				log.Println(err)
				continue
			}

			imageIndex[term] = append(imageIndex[term], types.ImagePosting{
				ImageUrl:      imageUrl,
				TermFrequency: frequency,
			})
		}
	}

	if err = db.AddImageIndex(imageIndex); err != nil {
		log.Println(err)
	}

	//add document
	document := types.Document{
		NormUrl: link,
		Length:  len(wordMap),
		Title:   title,
	}
	err = db.AddDocument(document)
	if err != nil {
		log.Println(err)
	}

	//add wordmap/index
	index := types.InvertedIndex{}
	for word, score := range wordMap {
		index[word] = types.Posting{
			TermFrequency: score,
			NormUrl:       link,
		}
	}

	err = db.AddIndex(index)
	if err != nil {
		log.Println(err)
	}

	end := time.Now()

	log.Printf("time taken ms: %v", end.UnixMilli()-start.UnixMilli())
}

func Start(ctx context.Context, db *database.DataBase) {
	qlen, err := db.UrlQueueLength()
	if err != nil {
		panic(err)
	}
	for qlen > 0 {
		select {
		case <-ctx.Done():
			fmt.Println("Crawler stopping")
			return
		default:
			CrawlJob(db)
		}
		qlen, err = db.UrlQueueLength()
		if err != nil {
			panic(err)
		}
	}
	if qlen == 0 {
		fmt.Println("queue is empty")
	}
}
