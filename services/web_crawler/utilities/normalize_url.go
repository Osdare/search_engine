package utilities

import (
	"fmt"
	"net/url"
)

func NormalizeLink(base string, link string) (string, error) {
	baseURL, err := url.Parse(base)
	if err != nil {
		return "", err
	}

	ref, err := url.Parse(link)
	if err != nil {
		return "", err
	}

	absURL := baseURL.ResolveReference(ref)

	if absURL.Scheme == "" || absURL.Host == "" {
		return "", nil
	}

	return absURL.String(), nil
}

func NormalizeUrlSlice(base string, rawUrls []string) ([]string, error) {
	normalizedUrls := make([]string, 0)

	for _, rawUrl := range rawUrls {
		normUrl, err := NormalizeLink(base, rawUrl)
		if err != nil {
			return normalizedUrls, fmt.Errorf("could not normalize url: %v %v", rawUrl, err)
		}

		if normUrl != "" {
			normalizedUrls = append(normalizedUrls, normUrl)
		}
	}

	return normalizedUrls, nil
}
