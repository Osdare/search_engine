package utilities

import (
	"fmt"
	"net/url"
	"strings"
)

func GetDomainFromUrl(rawUrl string) (string, error) {
	u, err := url.Parse(rawUrl)

	if err != nil {
		return "", fmt.Errorf("failed to parse URL: %w", err)
	}

	if u.Scheme == "http" {
		u.Host = strings.Replace(u.Host, ":80", "", 1)
	}

	if u.Scheme == "https" {
		u.Host = strings.Replace(u.Host, ":443", "", 1)
	}

	u.Host = strings.ToLower(u.Host)
	u.Host = strings.TrimPrefix(u.Host, "www.")

	return u.Host, nil
}
