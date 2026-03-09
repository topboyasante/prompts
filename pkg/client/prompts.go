package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

type Prompt struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
}

type PromptVersion struct {
	ID      string `json:"id"`
	Version string `json:"version"`
}

func (c *Client) CreatePrompt(name, description string, tags []string) (*Prompt, int, error) {
	resp, err := c.Do(http.MethodPost, "/prompts", map[string]any{
		"name":        name,
		"description": description,
		"tags":        tags,
	}, true)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		body, _ := io.ReadAll(resp.Body)
		return nil, resp.StatusCode, fmt.Errorf("create prompt failed: %s", string(body))
	}

	if resp.StatusCode == http.StatusConflict {
		return nil, resp.StatusCode, nil
	}

	var p Prompt
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, resp.StatusCode, err
	}
	return &p, resp.StatusCode, nil
}

func (c *Client) UploadVersion(promptID, version string, tarball []byte) error {
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	_ = w.WriteField("version", version)
	part, err := w.CreateFormFile("tarball", "prompt.tar.gz")
	if err != nil {
		return err
	}
	if _, err := part.Write(tarball); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, c.BaseURL+"/prompts/"+promptID+"/versions", &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload version failed: %s", string(body))
	}
	return nil
}

func (c *Client) SearchPrompts(query string) ([]Prompt, error) {
	resp, err := c.Do(http.MethodGet, "/prompts?q="+url.QueryEscape(query), nil, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("search failed: %s", string(body))
	}

	var out struct {
		Items []Prompt `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c *Client) GetPrompt(owner, name string) (*Prompt, error) {
	resp, err := c.Do(http.MethodGet, "/prompts/"+owner+"/"+name, nil, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get prompt failed: %s", strings.TrimSpace(string(body)))
	}
	var out struct {
		Prompt Prompt `json:"prompt"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return &out.Prompt, nil
}

func (c *Client) GetVersions(owner, name string) ([]PromptVersion, error) {
	resp, err := c.Do(http.MethodGet, "/prompts/"+owner+"/"+name+"/versions", nil, false)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list versions failed: %s", string(body))
	}
	var out struct {
		Items []PromptVersion `json:"items"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	return out.Items, nil
}

func (c *Client) DownloadVersion(owner, name, version string) ([]byte, error) {
	endpoint := c.BaseURL + "/prompts/" + owner + "/" + name + "/versions/" + version + "/download"
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}

	httpClient := *c.HTTPClient
	httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error { return nil }

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("download endpoint failed: %s", string(body))
	}
	location := resp.Header.Get("Location")
	if location == "" {
		return nil, fmt.Errorf("missing redirect location")
	}

	res2, err := c.HTTPClient.Get(location)
	if err != nil {
		return nil, err
	}
	defer res2.Body.Close()
	if res2.StatusCode >= http.StatusBadRequest {
		body, _ := io.ReadAll(res2.Body)
		return nil, fmt.Errorf("tarball download failed: %s", string(body))
	}

	return io.ReadAll(res2.Body)
}
