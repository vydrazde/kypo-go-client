package kypo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type TrainingDefinition struct {
	Id      int64  `json:"id" tfsdk:"id"`
	Content string `json:"content" tfsdk:"content"`
}

func (c *Client) GetTrainingDefinition(ctx context.Context, definitionID int64) (*TrainingDefinition, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/kypo-rest-training/api/v1/exports/training-definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/octet-stream")

	body, _, err := c.doRequestWithRetry(req, http.StatusOK, "training definition", definitionID)
	if err != nil {
		return nil, err
	}

	definition := TrainingDefinition{
		Id:      definitionID,
		Content: string(body),
	}

	return &definition, nil
}

func (c *Client) CreateTrainingDefinition(ctx context.Context, content string) (*TrainingDefinition, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/kypo-rest-training/api/v1/imports/training-definitions", c.Endpoint), strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	body, _, err := c.doRequestWithRetry(req, http.StatusOK, "training definition", "")
	if err != nil {
		return nil, err
	}

	id := struct {
		Id int64 `json:"id"`
	}{}

	err = json.Unmarshal(body, &id)
	if err != nil {
		return nil, err
	}

	definition := TrainingDefinition{
		Id:      id.Id,
		Content: content,
	}

	return &definition, nil
}

func (c *Client) DeleteTrainingDefinition(ctx context.Context, definitionID int64) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/kypo-rest-training/api/v1/training-definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return err
	}

	_, _, err = c.doRequestWithRetry(req, http.StatusOK, "training definition", definitionID)
	if err != nil {
		return err
	}

	return nil
}
