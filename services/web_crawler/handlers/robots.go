package handlers

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
	"web_crawler/types"
)

type RuleType int

const (
	RuleAllow RuleType = iota
	RuleDisallow
)

type Rule struct {
	Pattern    string
	ruleType   RuleType
	regex      *regexp.Regexp
	patternLen int
}

type RuleSet struct {
	Allow    []*Rule
	Disallow []*Rule
}

func newRule(pattern string, rtype RuleType) (*Rule, error) {
	var sb strings.Builder
	special := ".^$+?{}[]\\|()"
	for _, c := range pattern {
		switch c {
		case '*':
			sb.WriteString(".*")
		case '$':
			sb.WriteString("$")
		default:
			if strings.ContainsRune(special, c) {
				sb.WriteByte('\\')
			}
			sb.WriteRune(c)
		}
	}
	escaped := sb.String()
	regexPattern := "^" + escaped

	re, err := regexp.Compile(regexPattern)
	if err != nil {
		return nil, err
	}
	return &Rule{
		Pattern:    pattern,
		ruleType:   rtype,
		regex:      re,
		patternLen: len(pattern),
	}, nil
}

func newRuleSet(allowPatterns, disallowPatterns []string) (*RuleSet, error) {
	rs := &RuleSet{}
	for _, p := range allowPatterns {
		r, err := newRule(p, RuleAllow)
		if err != nil {
			return nil, err
		}
		rs.Allow = append(rs.Allow, r)
	}
	for _, p := range disallowPatterns {
		r, err := newRule(p, RuleDisallow)
		if err != nil {
			return nil, err
		}
		rs.Disallow = append(rs.Disallow, r)
	}
	return rs, nil
}

func (rs *RuleSet) isAllowed(rawurl string) bool {
	u, err := url.Parse(rawurl)
	if err != nil {
		return false
	}

	pathAndQuery := u.Path
	if u.RawQuery != "" {
		pathAndQuery += "?" + u.RawQuery
	}

	type match struct {
		ruleType   RuleType
		patternLen int
	}
	matches := []match{}

	for _, r := range rs.Allow {
		if r.regex.MatchString(pathAndQuery) {
			matches = append(matches, match{ruleType: RuleAllow, patternLen: r.patternLen})
		}
	}
	for _, r := range rs.Disallow {
		if r.regex.MatchString(pathAndQuery) {
			matches = append(matches, match{ruleType: RuleDisallow, patternLen: r.patternLen})
		}
	}

	if len(matches) == 0 {
		return true
	}

	best := matches[0]
	for _, m := range matches[1:] {
		if m.patternLen > best.patternLen {
			best = m
		} else if m.patternLen == best.patternLen {
			if m.ruleType == RuleAllow && best.ruleType == RuleDisallow {
				best = m
			}
		}
	}

	return best.ruleType == RuleAllow
}

type Reason int

const (
	ReasonCrawlDelay Reason = iota
	ReasonAllowed
	ReasonDisallowed
	ReasonFailed
)

func CanCrawl(rawUrl string, domain types.Domain) (bool, Reason, error) {
	rs, err := newRuleSet(domain.Allowed, domain.Disallowed)
	if err != nil {
		return false, ReasonFailed, err
	}

	if !rs.isAllowed(rawUrl) {
		return false, ReasonDisallowed, nil
	}

	if time.Now().Unix()-domain.LastCrawled < domain.CrawlDelay {
		return false, ReasonCrawlDelay, nil
	}

	return true, ReasonAllowed, nil
}

func downloadRobots(domainName string) ([]string, error) {
	url := domainName + "/robots.txt"
	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		return nil, fmt.Errorf("error from http request to %v %v", url, err)
	}

	userAgent := os.Getenv("USER_AGENT")

	req.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return make([]string, 0), fmt.Errorf("could not get robots from %v %v", domainName, err)
	}
	defer resp.Body.Close()

	bb, err := io.ReadAll(resp.Body)
	if err != nil {
		return make([]string, 0), fmt.Errorf("could not read bytes from body %v", err)
	}

	strings := make([]string, 0)
	s := ""
	for _, b := range bb {
		if b == '\n' {
			strings = append(strings, s)
			s = ""
		} else {
			s += string(b)
		}
	}

	return strings, nil
}

func robotsToDomain(domainName string, robotsLines []string) (types.Domain, error) {
    domain := types.Domain{Name: domainName}
    allow := []string{}
    disallow := []string{}
    isInUser := false
    userAgent := strings.ToLower(os.Getenv("USER_AGENT"))

    for _, raw := range robotsLines {
        line := strings.TrimSpace(raw)
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }

        lineLower := strings.ToLower(line)

        if strings.HasPrefix(lineLower, "user-agent: *") ||
            strings.HasPrefix(lineLower, "user-agent: "+userAgent) {
            isInUser = true
            continue
        }

        if isInUser && strings.HasPrefix(lineLower, "user-agent:") {
            break 
        }

        if !isInUser {
            continue
        }

        switch {
        case strings.HasPrefix(lineLower, "allow:"):
            s := strings.TrimSpace(strings.TrimPrefix(line, "Allow:"))
            allow = append(allow, s)

        case strings.HasPrefix(lineLower, "disallow:"):
            s := strings.TrimSpace(strings.TrimPrefix(line, "Disallow:"))
            disallow = append(disallow, s)

        case strings.HasPrefix(lineLower, "crawl-delay:"):
            parts := strings.Fields(line)
            if len(parts) > 1 {
                cd, err := strconv.ParseInt(parts[1], 10, 64)
                if err != nil {
                    return types.Domain{}, fmt.Errorf("invalid crawl-delay in line %q: %v", line, err)
                }
                domain.CrawlDelay = cd
            }
        }
    }

    domain.Allowed = allow
    domain.Disallowed = disallow
    domain.LastCrawled = time.Now().Unix()

    return domain, nil
}

func GetRobotsFromDomain(domainName string) (types.Domain, error) {
	u, err := url.Parse(domainName)
	if err != nil {
		return types.Domain{Name: u.Host}, fmt.Errorf("could not parse domain: %v Reason: %v", domainName, err)
	}

	lines, err := downloadRobots(domainName)
	if err != nil {
		return types.Domain{Name: u.Host}, err
	}

	return robotsToDomain(u.Host, lines)
}
