package plane

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	baseURL    string
	apiToken   string
	workspace  string
	httpClient *http.Client
}

func NewClient(baseURL, apiToken, workspace string) *Client {
	return &Client{
		baseURL:   baseURL,
		apiToken:  apiToken,
		workspace: workspace,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

func (c *Client) do(path string, out any) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Api-Key", c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("plane API %s: HTTP %d", path, resp.StatusCode)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *Client) Projects() ([]Project, error) {
	var result ListResponse[Project]
	err := c.do(fmt.Sprintf("/api/v1/workspaces/%s/projects/", c.workspace), &result)
	return result.Results, err
}

func (c *Client) Issues(projectID string) ([]Issue, error) {
	var result ListResponse[Issue]
	path := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/issues/?per_page=100", c.workspace, projectID)
	err := c.do(path, &result)
	return result.Results, err
}

func (c *Client) Issue(projectID, issueID string) (*Issue, error) {
	var issue Issue
	path := fmt.Sprintf("/api/v1/workspaces/%s/projects/%s/issues/%s/", c.workspace, projectID, issueID)
	err := c.do(path, &issue)
	return &issue, err
}
