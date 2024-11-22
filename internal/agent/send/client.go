package send

import (
	"github.com/gdyunin/metricol.git/internal/agent/fetch"
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"
)

type Client struct {
	storage  *fetch.Storage
	host     string
	basePath string
	client   http.Client
}

func NewClient(s *fetch.Storage, host string, port int) *Client {
	return &Client{
		storage:  s,
		host:     host + ":" + strconv.Itoa(port),
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

		res, err := c.client.Do(r)
		if err != nil {
			return err
		}
		defer func() { _ = res.Body.Close() }()

		if _, err := io.ReadAll(res.Body); err != nil {
			return err
		}
	}
	return nil
}

func newBasePath() string {
	return "/update/"
}
