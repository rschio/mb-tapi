package tapi

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"
)

// DefaultService is the default endpoint to tapi.
const DefaultService = "https://www.mercadobitcoin.net/tapi/v3/"

// Client is a API client.
type Client struct {
	service string
	apiID   string
	apiKey  string
	client  *http.Client
}

// NewClient creates a new client.
func NewClient(service, apiID, apiKey string, client *http.Client) *Client {
	c := &Client{
		service: service,
		apiID:   apiID,
		apiKey:  apiKey,
		client:  client,
	}
	if c.client == nil {
		c.client = http.DefaultClient
	}
	return c
}

// Nonce creates a unique value that always increase.
func (c *Client) Nonce() string {
	t := time.Now().UnixNano()
	nonce := strconv.FormatInt(t, 10)
	return nonce
}

// Hmac signs the msg.
func (c *Client) Hmac(msg string) string {
	mac := hmac.New(sha512.New, []byte(c.apiKey))
	mac.Write([]byte(msg))
	out := mac.Sum(nil)
	return hex.EncodeToString(out)
}
