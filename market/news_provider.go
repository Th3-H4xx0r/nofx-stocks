package market

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const newsAPIURL = "https://newsapi.org/v2/everything"

type Article struct {
	Source      struct{ Name string } `json:"source"`
	Author      string                `json:"author"`
	Title       string                `json:"title"`
	Description string                `json:"description"`
	URL         string                `json:"url"`
	PublishedAt time.Time             `json:"publishedAt"`
	Content     string                `json:"content"`
}

type NewsResponse struct {
	Status       string    `json:"status"`
	TotalResults int       `json:"totalResults"`
	Articles     []Article `json:"articles"`
}

// FetchNews fetches news for a query (e.g. symbol or keywords)
func FetchNews(query string, apiKey string, domains []string, limit int) ([]Article, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("NewsAPI key is required")
	}

	req, err := http.NewRequest("GET", newsAPIURL, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("q", query)
	q.Add("apiKey", apiKey)
	q.Add("language", "en")
	q.Add("sortBy", "publishedAt")
	q.Add("pageSize", fmt.Sprintf("%d", limit))

	if len(domains) > 0 {
		// domains=wsj.com,bloomberg.com
		// Check strings.Join
		q.Add("domains", "") // Placeholder, logic needed
	}

	// Simple fix for domains: join manually
	if len(domains) > 0 {
		domainStr := ""
		for i, d := range domains {
			if i > 0 {
				domainStr += ","
			}
			domainStr += d
		}
		q.Add("domains", domainStr)
	}

	req.URL.RawQuery = q.Encode()

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("NewsAPI error: %s", string(body))
	}

	var result NewsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Articles, nil
}
