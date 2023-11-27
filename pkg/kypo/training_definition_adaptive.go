package kypo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type TrainingDefinitionAdaptive struct {
	Id      int64  `json:"id" tfsdk:"id"`
	Content string `json:"content" tfsdk:"content"`
}

// GetTrainingDefinitionAdaptive reads the adaptive definition given by definitionID.
func (c *Client) GetTrainingDefinitionAdaptive(ctx context.Context, definitionID int64) (*TrainingDefinitionAdaptive, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/kypo-adaptive-training/api/v1/exports/training-definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("accept", "application/octet-stream")

	body, _, err := c.doRequestWithRetry(req, http.StatusOK, "training definition adaptive", definitionID)
	if err != nil {
		return nil, err
	}

	definition := TrainingDefinitionAdaptive{
		Id:      definitionID,
		Content: string(body),
	}

	return &definition, nil
}

// CreateTrainingDefinitionAdaptive imports a JSON string content as a adaptive training definition.
// The JSON string must be a previously exported adaptive training definition.
// Since KYPO returns an answer with a definition in a different format than the exported definition,
// only the Id is read and the input content is set as the returned TrainingDefinition.Content.
func (c *Client) CreateTrainingDefinitionAdaptive(ctx context.Context, content string) (*TrainingDefinitionAdaptive, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/kypo-adaptive-training/api/v1/imports/training-definitions", c.Endpoint), strings.NewReader(content))
	if err != nil {
		return nil, err
	}

	body, _, err := c.doRequestWithRetry(req, http.StatusOK, "training definition adaptive", "")
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

	definition := TrainingDefinitionAdaptive{
		Id:      id.Id,
		Content: content,
	}

	return &definition, nil
}

// DeleteTrainingDefinitionAdaptive deletes the adaptive definition given by definitionID.
func (c *Client) DeleteTrainingDefinitionAdaptive(ctx context.Context, definitionID int64) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/kypo-adaptive-training/api/v1/training-definitions/%d", c.Endpoint, definitionID), nil)
	if err != nil {
		return err
	}

	_, _, err = c.doRequestWithRetry(req, http.StatusOK, "training definition adaptive", definitionID)
	if err != nil {
		return err
	}

	return nil
}
