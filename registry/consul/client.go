package consul

// consul.Client makes it easy to communicate with the consul API

type Client struct {
	url string
}

func NewClient(url string) *Client {
	return &Client{url: url}
}

func (c *Client) Register() error {
	return nil
}
