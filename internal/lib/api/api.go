package api

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	ErrInvalidStatusCode = errors.New("invalid status code")
)

// GetRedirect returns the final URL after redirection
func GetRedirect(url string) (string, error) {
	const op = "api.GetRedirect"

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusFound {
		return "", fmt.Errorf("%s: %w: %s", op, ErrInvalidStatusCode, resp.Status)
	}

	defer func() { _ = resp.Body.Close() }()

	return resp.Header.Get("Location"), nil
}

func DeleteURL(url string) (int, error) {
	const op = "api.DeleteURL"

	client := &http.Client{}
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	return resp.StatusCode, nil
}
