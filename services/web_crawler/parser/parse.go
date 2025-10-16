package parser

import (
	"slices"
	"strings"
	"utils"
	"web_crawler/consts"
	"web_crawler/types"

	"github.com/reiver/go-porterstemmer"
	"golang.org/x/net/html"
)

// Single pass over the html
func ParseBody(normUrl string, body *html.Node) (title string, rawUrls []string, images []types.Image, wordMap map[string]int) {
	wordMap = make(map[string]int)
	images = make([]types.Image, 0)
	rawUrls = make([]string, 0)

	var f func(*html.Node)
	f = func(n *html.Node) {
		switch n.Type {
		case html.TextNode:
			w := strings.SplitSeq(n.Data, " ")
			for word := range w {
				//normalizing and stemming
				word = strings.ToLower(word)
				word = strings.TrimSpace(word)
				word = utils.RemovePunctuation(word)

				if len(word) < 3 {
					continue
				}

				stem := porterstemmer.StemWithoutLowerCasing([]rune(word))

				if len(stem) >= 2 &&
					len(stem) <= 32 &&
					!slices.Contains(consts.StopWords, word) &&
					utils.IsAlphanumeric(string(stem)) {
					wordMap[string(stem)]++
				}
			}
		case html.ElementNode:
			if n.Data == "title" && n.FirstChild != nil {
				title = n.FirstChild.Data
			} else if n.Data == "a" {
				for _, attr := range n.Attr {
					if attr.Key == "href" {
						rawUrls = append(rawUrls, attr.Val)
					}
				}
			} else if n.Data == "img" {
				image := types.Image{
					PageUrl: normUrl,
				}
				for _, attr := range n.Attr {
					if attr.Key == "src" {
						image.ImageUrl = attr.Val
					}
					if attr.Key == "alt" {
						image.Text += attr.Val
					}
				}

				if fig := findParentFigure(n); fig != nil {
					caption := findFigcaption(fig)
					image.Text += " " + caption
				}

				images = append(images, image)
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}
	f(body)
	return title, rawUrls, images, wordMap
}

// climb up to parent <figure>
func findParentFigure(n *html.Node) *html.Node {
	for p := n.Parent; p != nil; p = p.Parent {
		if p.Type == html.ElementNode && p.Data == "figure" {
			return p
		}
	}
	return nil
}

// recursive search for <figcaption> inside a node
func findFigcaption(n *html.Node) string {
	if n.Type == html.ElementNode && n.Data == "figcaption" {
		if n.FirstChild != nil && n.FirstChild.Type == html.TextNode {
			return strings.TrimSpace(n.FirstChild.Data)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if caption := findFigcaption(c); caption != "" {
			return caption
		}
	}
	return ""
}
