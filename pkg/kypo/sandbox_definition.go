package kypo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type SandboxDefinition struct {
	Id        int64     `json:"id" tfsdk:"id"`
	Url       string    `json:"url" tfsdk:"url"`
	Name      string    `json:"name" tfsdk:"name"`
	Rev       string    `json:"rev" tfsdk:"rev"`
	CreatedBy UserModel `json:"created_by" tfsdk:"created_by"`
}

type SandboxDefinitionRequest struct {
	Url string `json:"url"`
	Rev string `json:"rev"`
}

// GetSandboxDefinition reads the given sandbox definition.
func (c *Client) GetSandboxDefinition(ctx context.Context, definitionID int64) (*SandboxDefinition, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return nil, err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	definition := SandboxDefinition{}

	if status == http.StatusNotFound {
		return nil, &ErrNotFound{ResourceName: "sandbox definition", Identifier: strconv.FormatInt(definitionID, 10)}
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
	}

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
	requestBody, err := json.Marshal(SandboxDefinitionRequest{url, rev})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/definitions", c.Endpoint), strings.NewReader(string(requestBody)))
	if err != nil {
		return nil, err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if status != http.StatusCreated {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
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

	body, status, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if status != http.StatusNoContent && status != http.StatusNotFound {
		return fmt.Errorf("status: %d, body: %s", status, body)
	}

	return nil
}
