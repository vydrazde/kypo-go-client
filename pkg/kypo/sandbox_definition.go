package kypo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type SandboxDefinition struct {
	Id        int64  `json:"id" tfsdk:"id"`
	Url       string `json:"url" tfsdk:"url"`
	Name      string `json:"name" tfsdk:"name"`
	Rev       string `json:"rev" tfsdk:"rev"`
	CreatedBy User   `json:"created_by" tfsdk:"created_by"`
}

type sandboxDefinitionRequest struct {
	Url string `json:"url"`
	Rev string `json:"rev"`
}

// GetSandboxDefinition reads the given sandbox definition.
func (c *Client) GetSandboxDefinition(ctx context.Context, definitionID int64) (*SandboxDefinition, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return nil, err
	}

	body, _, err := c.doRequestWithRetry(req, http.StatusOK, "sandbox definition", definitionID)
	if err != nil {
		return nil, err
	}

	definition := SandboxDefinition{}
	err = json.Unmarshal(body, &definition)
	if err != nil {
		return nil, err
	}

	return &definition, nil
}

// CreateSandboxDefinition creates a sandbox definition.
// The `url` must be a URL to a GitLab repository where the sandbox definition is hosted.
// The `rev` specifies the Git revision to be used.
func (c *Client) CreateSandboxDefinition(ctx context.Context, url, rev string) (*SandboxDefinition, error) {
	requestBody, err := json.Marshal(sandboxDefinitionRequest{url, rev})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/definitions", c.Endpoint), strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, err
	}

	body, _, err := c.doRequestWithRetry(req, http.StatusCreated, "sandbox definition", "")
	if err != nil {
		return nil, err
	}

	definition := SandboxDefinition{}
	err = json.Unmarshal(body, &definition)
	if err != nil {
		return nil, err
	}

	return &definition, nil
}

// DeleteSandboxDefinition deletes the given sandbox definition.
func (c *Client) DeleteSandboxDefinition(ctx context.Context, definitionID int64) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return err
	}

	_, _, err = c.doRequestWithRetry(req, http.StatusNoContent, "sandbox definition", definitionID)
	if err != nil {
		return err
	}

	return nil
}
