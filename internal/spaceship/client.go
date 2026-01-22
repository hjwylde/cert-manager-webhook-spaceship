package spaceship

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const apiBaseURL = "https://spaceship.dev/api/v1/"

type Client struct {
	client *http.Client

	BaseURL   *url.URL
	ApiSecret string
	ApiKey    string

	DNSRecords *DNSRecordsService
}

func NewClient() (*Client, error) {
	uri, err := url.Parse(apiBaseURL)
	if err != nil {
		return nil, err
	}

	c := &Client{client: &http.Client{}, BaseURL: uri}

	s := service{c}
	c.DNSRecords = (*DNSRecordsService)(&s)

	return c, nil
}

func (c *Client) NewRequest(method, uriStr string, body interface{}) (*http.Request, error) {
	if !strings.HasSuffix(c.BaseURL.Path, "/") {
		return nil, fmt.Errorf("baseURL must have a trailing slash, but %q does not", c.BaseURL)
	}

	uri, err := c.BaseURL.Parse(uriStr)
	if err != nil {
		return nil, err
	}

	var buf io.ReadWriter
	if body != nil {
		buf = &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.SetEscapeHTML(false)
		err := enc.Encode(body)
		if err != nil {
			return nil, err
		}
	}

	req, err := http.NewRequest(method, uri.String(), buf)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.ApiKey != "" {
		req.Header.Set("X-Api-Key", c.ApiKey)
	}
	if c.ApiSecret != "" {
		req.Header.Set("X-Api-Secret", c.ApiSecret)
	}

	return req, nil
}

func (c *Client) Do(ctx context.Context, request *http.Request, value interface{}) (*http.Response, error) {
	resp, err := c.do(ctx, c.client, request)
	if resp != nil {
		//goland:noinspection GoUnhandledErrorResult
		defer resp.Body.Close()
	}
	if err != nil {
		return resp, err
	}

	if c := resp.StatusCode; c == http.StatusAccepted {
		return resp, &HTTPAcceptedError{}
	} else if c < 200 || 300 <= c {
		return resp, &HTTPStatusError{resp.Status, resp.StatusCode}
	}

	switch value := value.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(value, resp.Body)
	default:
		err := json.NewDecoder(resp.Body).Decode(value)
		if err == io.EOF {
			err = nil
		}
	}

	return resp, err
}

func (c *Client) do(ctx context.Context, caller *http.Client, request *http.Request) (*http.Response, error) {
	resp, err := caller.Do(request.WithContext(ctx))
	if err != nil {
		select {
		case <-ctx.Done():
			return resp, ctx.Err()
		default:
		}

		return resp, err
	}

	return resp, nil
}

type HTTPAcceptedError struct {
}

func (e *HTTPAcceptedError) Error() string {
	return http.StatusText(http.StatusAccepted)
}

type HTTPStatusError struct {
	Status     string
	StatusCode int
}

func (e *HTTPStatusError) Error() string {
	return e.Status
}

type service struct {
	client *Client
}
