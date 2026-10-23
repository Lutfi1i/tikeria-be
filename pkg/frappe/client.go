package frappe

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type FrappeClient interface {
	GetDoc(doctype, name string) (map[string]interface{}, error)
	CreateDoc(doctype string, data map[string]interface{}) (map[string]interface{}, error)
	UpdateDoc(doctype, name string, data map[string]interface{}) (map[string]interface{}, error)
	GetList(doctype string, filters map[string]interface{}, fields []string) ([]map[string]interface{}, error)
}

type Client struct {
	BaseURL    string
	APIKey     string
	APISecret  string
	HTTPClient *http.Client
}

func NewClient(baseURL, apiKey, apiSecret string) *Client {
	return &Client{
		BaseURL:   strings.TrimRight(baseURL, "/"),
		APIKey:    apiKey,
		APISecret: apiSecret,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) addHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.APIKey != "" && c.APISecret != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s:%s", c.APIKey, c.APISecret))
	}
}

func (c *Client) GetDoc(doctype, name string) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/api/resource/%s/%s", c.BaseURL, url.PathEscape(doctype), url.PathEscape(name))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("frappe error (%d): %s", resp.StatusCode, string(body))
	}

	var res struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) CreateDoc(doctype string, data map[string]interface{}) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/api/resource/%s", c.BaseURL, url.PathEscape(doctype))
	bodyBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("frappe create error (%d): %s", resp.StatusCode, string(body))
	}

	var res struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) UpdateDoc(doctype, name string, data map[string]interface{}) (map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/api/resource/%s/%s", c.BaseURL, url.PathEscape(doctype), url.PathEscape(name))
	bodyBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("frappe update error (%d): %s", resp.StatusCode, string(body))
	}

	var res struct {
		Data map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return res.Data, nil
}

func (c *Client) GetList(doctype string, filters map[string]interface{}, fields []string) ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/api/resource/%s", c.BaseURL, url.PathEscape(doctype))
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}

	q := u.Query()
	if len(filters) > 0 {
		filterJSON, err := json.Marshal(filters)
		if err == nil {
			q.Set("filters", string(filterJSON))
		}
	}
	if len(fields) > 0 {
		fieldsJSON, err := json.Marshal(fields)
		if err == nil {
			q.Set("fields", string(fieldsJSON))
		}
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	c.addHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("frappe list error (%d): %s", resp.StatusCode, string(body))
	}

	var res struct {
		Data []map[string]interface{} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return nil, err
	}
	return res.Data, nil
}
