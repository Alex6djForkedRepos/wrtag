package musicbrainz

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/time/rate"
)

type CAAClient struct {
	httpClient *http.Client
	limiter    *rate.Limiter
}

func NewCAAClient(httpClient *http.Client) *CAAClient {
	return &CAAClient{
		httpClient: httpClient,
		limiter:    rate.NewLimiter(rate.Inf, 1), // https://wiki.musicbrainz.org/Cover_Art_Archive/API#Rate_limiting_rules
	}
}

func (c *CAAClient) request(ctx context.Context, r *http.Request, dest any) error {
	if err := c.limiter.Wait(ctx); err != nil {
		return fmt.Errorf("wait: %w", err)
	}

	log.Printf("making caa request %s", r.URL)

	r = r.WithContext(ctx)
	resp, err := c.httpClient.Do(r)
	if err != nil {
		return fmt.Errorf("make caa request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		return fmt.Errorf("caa returned non 2xx: %w", StatusError(resp.StatusCode))
	}
	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("decode caa response: %w", err)
	}
	return nil
}

func (c *CAAClient) GetCoverURL(ctx context.Context, release *Release) (string, error) {
	// try release first
	if release.CoverArtArchive.Front {
		url, err := c.getCoverURL(ctx, joinPath(caaBase, "release", release.ID))
		if err != nil {
			return "", fmt.Errorf("try release: %w", err)
		}
		if url != "" {
			return url, nil
		}
	}

	// fall back to release group
	url, err := c.getCoverURL(ctx, joinPath(caaBase, "release-group", release.ReleaseGroup.ID))
	if err != nil {
		return "", fmt.Errorf("try release group: %w", err)
	}
	if url != "" {
		return url, nil
	}
	return "", nil
}

func (c *CAAClient) getCoverURL(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	var caa caaResponse
	err = c.request(ctx, req, &caa)
	if se := StatusError(0); errors.As(err, &se) && se == http.StatusNotFound {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("make caa release request: %w", err)
	}

	for _, img := range caa.Images {
		if img.Front {
			return img.Image, nil
		}
	}
	return "", nil
}

type caaResponse struct {
	Release string `json:"release"`
	Images  []struct {
		Approved   bool     `json:"approved"`
		Back       bool     `json:"back"`
		Comment    string   `json:"comment"`
		Edit       int      `json:"edit"`
		Front      bool     `json:"front"`
		ID         any      `json:"id"`
		Image      string   `json:"image"`
		Types      []string `json:"types"`
		Thumbnails struct {
			Num250  string `json:"250"`
			Num500  string `json:"500"`
			Num1200 string `json:"1200"`
			Large   string `json:"large"`
			Small   string `json:"small"`
		} `json:"thumbnails"`
	} `json:"images"`
}
