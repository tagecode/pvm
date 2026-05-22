package download

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Client struct {
	HTTP *http.Client
}

func NewClient() *Client {
	return &Client{HTTP: defaultHTTPClient()}
}

func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Minute,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   15 * time.Second,
			ResponseHeaderTimeout: 60 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
		},
	}
}

func (c *Client) Fetch(ctx context.Context, url, dest string) error {
	if c.HTTP == nil {
		c.HTTP = defaultHTTPClient()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTP.Do(req)
	if err != nil {
		return fmt.Errorf("download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return err
	}

	part := dest + ".part"
	f, err := os.Create(part)
	if err != nil {
		return err
	}

	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(part)
		return err
	}
	if err := f.Close(); err != nil {
		os.Remove(part)
		return err
	}

	if err := os.RemoveAll(dest); err != nil {
		return err
	}
	return os.Rename(part, dest)
}
