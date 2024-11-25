package send

import (
	"github.com/gdyunin/metricol.git/internal/agent/fetch"
	"net/http"
	"net/url"
	"path"
)

type Client struct {
	storage  *fetch.Storage
	host     string
	basePath string
	client   http.Client
}

func NewClient(s *fetch.Storage, address string) *Client {
	return &Client{
		storage:  s,
		host:     address,
		basePath: newBasePath(),
		client:   http.Client{},
	}
}

func (c *Client) Send() error {
	for _, m := range c.storage.Metrics() {
		u := url.URL{
			Scheme: "http",
			Host:   c.host,
			Path:   path.Join(c.basePath, path.Join(string(m.Type()), m.Name(), m.Value())),
		}

		r, err := http.NewRequest(http.MethodPost, u.String(), nil)
		if err != nil {
			return err
		}
		r.Header.Set("Content-Type", "text/plain")

		res, _ := c.client.Do(r)
		if res != nil {
			_ = res.Body.Close()
		}
	}
	return nil
}

func newBasePath() string {
	return "/update/"
}
