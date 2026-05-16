package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

const (
	raindropsURL     = "https://api.raindrop.io/rest/v1/raindrops/0"
	raindropsPerPage = 50
)

type Raindrop struct {
	ID        int
	Title     string
	Link      string
	Domain    string
	CreatedAt time.Time
}

type raindropResponse struct {
	ID      int    `json:"_id"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	Domain  string `json:"domain"`
	Created string `json:"created"`
}

type raindropsResponse struct {
	Result bool               `json:"result"`
	Items  []raindropResponse `json:"items"`
	Count  int                `json:"count"`
}

func FetchAllRaindrops(client *http.Client, accessToken string, search string) ([]Raindrop, error) {
	var all []Raindrop
	for page := 0; ; page++ {
		items, err := fetchRaindropsPage(client, accessToken, page, search)
		if err != nil {
			return nil, err
		}
		all = append(all, items...)
		if len(items) < raindropsPerPage {
			break
		}
	}
	return all, nil
}

func fetchRaindropsPage(client *http.Client, accessToken string, page int, search string) ([]Raindrop, error) {
	endpoint, err := url.Parse(raindropsURL)
	if err != nil {
		return nil, fmt.Errorf("parse raindrops URL: %w", err)
	}
	query := endpoint.Query()
	query.Set("page", strconv.Itoa(page))
	query.Set("perpage", strconv.Itoa(raindropsPerPage))
	if search != "" {
		query.Set("search", search)
	}
	endpoint.RawQuery = query.Encode()

	req, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("create raindrops request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request raindrops: %w", err)
	}
	defer resp.Body.Close()

	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read raindrops response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, responseError("request raindrops", resp.Status, responseBody)
	}

	var result raindropsResponse
	if err := json.Unmarshal(responseBody, &result); err != nil {
		return nil, fmt.Errorf("decode raindrops response: %w", err)
	}
	if !result.Result {
		return nil, errors.New("raindrops response result is false")
	}
	return result.raindrops(), nil
}

func (r raindropsResponse) raindrops() []Raindrop {
	items := make([]Raindrop, 0, len(r.Items))
	for _, item := range r.Items {
		items = append(items, item.raindrop())
	}
	return items
}

func (r raindropResponse) raindrop() Raindrop {
	return Raindrop{
		ID:        r.ID,
		Title:     r.Title,
		Link:      r.Link,
		Domain:    r.Domain,
		CreatedAt: parseRaindropCreatedAt(r.Created),
	}
}

func parseRaindropCreatedAt(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	createdAt, err := time.Parse(time.RFC3339, value)
	if err == nil {
		return createdAt
	}
	createdAt, err = time.Parse("2006-01-02T15:04:05.000Z", value)
	if err == nil {
		return createdAt
	}
	return time.Time{}
}
