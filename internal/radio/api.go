// Package radio provides a client for the radio-browser.info community API.
package radio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	apiBase    = "https://de1.api.radio-browser.info/json"
	userAgent  = "spore/1.0 (https://github.com/spore-player/spore)"
	apiTimeout = 10 * time.Second
)

var httpClient = &http.Client{Timeout: apiTimeout}

// Station represents a radio station from radio-browser.info.
type Station struct {
	StationUUID string `json:"stationuuid"`
	Name        string `json:"name"`
	URL         string `json:"url_resolved"`
	Homepage    string `json:"homepage"`
	Country     string `json:"country"`
	Language    string `json:"language"`
	Tags        string `json:"tags"`
	Codec       string `json:"codec"`
	Bitrate     int    `json:"bitrate"`
	Votes       int    `json:"votes"`
	Clickcount  int    `json:"clickcount"`
	Favicon     string `json:"favicon"`
}

// TagName returns the station name, falling back to the URL host.
func (s Station) TagList() []string {
	if s.Tags == "" {
		return nil
	}
	parts := strings.Split(s.Tags, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// SearchParams holds options for station search.
type SearchParams struct {
	Name     string
	Tag      string
	Country  string
	Language string
	Codec    string
	Limit    int
	Offset   int
	Order    string // name, votes, clickcount, bitrate
	Reverse  bool
}

func doGet(ctx context.Context, endpoint string, params url.Values) ([]byte, error) {
	u := apiBase + endpoint
	if len(params) > 0 {
		u += "?" + params.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", userAgent)
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("radio api: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("radio api: HTTP %s", resp.Status)
	}
	var buf []byte
	tmp := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(tmp)
		buf = append(buf, tmp[:n]...)
		if err != nil {
			break
		}
	}
	return buf, nil
}

// Search queries stations by name/tag/country/language.
func Search(ctx context.Context, p SearchParams) ([]Station, error) {
	params := url.Values{}
	if p.Name != "" {
		params.Set("name", p.Name)
	}
	if p.Tag != "" {
		params.Set("tag", p.Tag)
	}
	if p.Country != "" {
		params.Set("country", p.Country)
	}
	if p.Language != "" {
		params.Set("language", p.Language)
	}
	if p.Codec != "" {
		params.Set("codec", p.Codec)
	}
	limit := p.Limit
	if limit <= 0 {
		limit = 50
	}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("offset", fmt.Sprintf("%d", p.Offset))
	order := p.Order
	if order == "" {
		order = "votes"
	}
	params.Set("order", order)
	if p.Reverse {
		params.Set("reverse", "true")
	}
	params.Set("hidebroken", "true")

	body, err := doGet(ctx, "/stations/search", params)
	if err != nil {
		return nil, err
	}
	var stations []Station
	if err := json.Unmarshal(body, &stations); err != nil {
		return nil, fmt.Errorf("radio api decode: %w", err)
	}
	return stations, nil
}

// TopVoted returns the top voted stations.
func TopVoted(ctx context.Context, limit int) ([]Station, error) {
	if limit <= 0 {
		limit = 50
	}
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("hidebroken", "true")
	body, err := doGet(ctx, "/stations/topvote", params)
	if err != nil {
		return nil, err
	}
	var stations []Station
	return stations, json.Unmarshal(body, &stations)
}

// TopClicked returns the most clicked stations recently.
func TopClicked(ctx context.Context, limit int) ([]Station, error) {
	if limit <= 0 {
		limit = 50
	}
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("hidebroken", "true")
	body, err := doGet(ctx, "/stations/topclick", params)
	if err != nil {
		return nil, err
	}
	var stations []Station
	return stations, json.Unmarshal(body, &stations)
}

// ByTag returns stations matching a tag exactly.
func ByTag(ctx context.Context, tag string, limit int) ([]Station, error) {
	return Search(ctx, SearchParams{Tag: tag, Limit: limit, Order: "votes", Reverse: true})
}

// Tags returns the most popular tags (genre/category names) from the API.
func Tags(ctx context.Context, limit int) ([]TagEntry, error) {
	if limit <= 0 {
		limit = 100
	}
	params := url.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	params.Set("order", "stationcount")
	params.Set("reverse", "true")
	body, err := doGet(ctx, "/tags", params)
	if err != nil {
		return nil, err
	}
	var tags []TagEntry
	return tags, json.Unmarshal(body, &tags)
}

// TagEntry is a tag with a station count.
type TagEntry struct {
	Name         string `json:"name"`
	StationCount int    `json:"stationcount"`
}

// Click registers a click/play event with the API (for statistics).
func Click(ctx context.Context, stationUUID string) {
	doGet(ctx, "/url/"+stationUUID, nil) //nolint:errcheck
}
