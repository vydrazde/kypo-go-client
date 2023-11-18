package kypo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"golang.org/x/exp/slices"
)

type SandboxAllocationUnit struct {
	Id                int64          `json:"id" tfsdk:"id"`
	PoolId            int64          `json:"pool_id" tfsdk:"pool_id"`
	AllocationRequest SandboxRequest `json:"allocation_request" tfsdk:"allocation_request"`
	CleanupRequest    SandboxRequest `json:"cleanup_request" tfsdk:"cleanup_request"`
	CreatedBy         UserModel      `json:"created_by" tfsdk:"created_by"`
	Locked            bool           `json:"locked" tfsdk:"locked"`
}

type SandboxRequest struct {
	Id               int64    `json:"id" tfsdk:"id"`
	AllocationUnitId int64    `json:"allocation_unit_id" tfsdk:"allocation_unit_id"`
	Created          string   `json:"created" tfsdk:"created"`
	Stages           []string `json:"stages" tfsdk:"stages"`
}

type SandboxRequestStageOutput struct {
	Page       int64  `json:"page" tfsdk:"page"`
	PageSize   int64  `json:"page_size" tfsdk:"page_size"`
	PageCount  int64  `json:"page_count" tfsdk:"page_count"`
	Count      int64  `json:"count" tfsdk:"line_count"`
	TotalCount int64  `json:"total_count" tfsdk:"total_count"`
	Result     string `json:"result" tfsdk:"result"`
}

type sandboxRequestStageOutputRaw struct {
	Page       int64        `json:"page"`
	PageSize   int64        `json:"page_size"`
	PageCount  int64        `json:"page_count"`
	Count      int64        `json:"count"`
	TotalCount int64        `json:"total_count"`
	Results    []outputLine `json:"results"`
}

type outputLine struct {
	Content string `json:"content"`
}

func (c *Client) GetSandboxAllocationUnit(ctx context.Context, unitId int64) (*SandboxAllocationUnit, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/sandbox-allocation-units/%d", c.Endpoint, unitId), nil)
	if err != nil {
		return nil, err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	allocationUnit := SandboxAllocationUnit{}

	if status == http.StatusNotFound {
		return nil, &ErrNotFound{ResourceName: "sandbox allocation unit", Identifier: strconv.FormatInt(unitId, 10)}
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
	}

	err = json.Unmarshal(body, &allocationUnit)
	if err != nil {
		return nil, err
	}

	return &allocationUnit, nil
}

func (c *Client) CreateSandboxAllocationUnits(ctx context.Context, poolId, count int64) ([]SandboxAllocationUnit, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/pools/%d/sandbox-allocation-units?count=%d", c.Endpoint, poolId, count), nil)
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

	var allocationUnit []SandboxAllocationUnit
	err = json.Unmarshal(body, &allocationUnit)
	if err != nil {
		return nil, err
	}

	return allocationUnit, nil
}

func (c *Client) CreateSandboxAllocationUnitAwait(ctx context.Context, poolId int64, pollTime time.Duration) (*SandboxAllocationUnit, error) {
	units, err := c.CreateSandboxAllocationUnits(ctx, poolId, 1)
	if err != nil {
		return nil, err
	}
	if len(units) != 1 {
		return nil, fmt.Errorf("expected one allocation unit to be created, got %d instead", len(units))
	}
	unit := units[0]
	request, err := c.PollRequestFinished(ctx, unit.Id, pollTime, "allocation")
	if err != nil {
		return nil, err
	}
	unit.AllocationRequest = *request
	return &unit, err
}

func (c *Client) CreateSandboxCleanupRequest(ctx context.Context, unitId int64) (*SandboxRequest, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/sandbox-allocation-units/%d/cleanup-request", c.Endpoint, unitId), nil)
	if err != nil {
		return nil, err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	if status == http.StatusNotFound {
		return nil, &ErrNotFound{ResourceName: "sandbox allocation unit", Identifier: strconv.FormatInt(unitId, 10)}
	}

	if status != http.StatusCreated {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
	}

	sandboxRequest := SandboxRequest{}
	err = json.Unmarshal(body, &sandboxRequest)
	if err != nil {
		return nil, err
	}

	return &sandboxRequest, nil
}

func (c *Client) PollRequestFinished(ctx context.Context, unitId int64, pollTime time.Duration, requestType string) (*SandboxRequest, error) {
	ticker := time.NewTicker(pollTime)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/sandbox-allocation-units/%d/%s-request", c.Endpoint, unitId, requestType), nil)
			if err != nil {
				return nil, err
			}

			body, status, err := c.doRequest(req)
			if err != nil {
				return nil, err
			}

			if status == http.StatusNotFound {
				return nil, &ErrNotFound{ResourceName: "sandbox request", Identifier: strconv.FormatInt(unitId, 10)}
			}

			if status != http.StatusOK {
				return nil, fmt.Errorf("status: %d, body: %s", status, body)
			}
			sandboxRequest := SandboxRequest{}
			err = json.Unmarshal(body, &sandboxRequest)
			if err != nil {
				return nil, err
			}

			if !slices.Contains(sandboxRequest.Stages, "RUNNING") && !slices.Contains(sandboxRequest.Stages, "IN_QUEUE") {
				return &sandboxRequest, nil
			}
		}
	}
}

func (c *Client) CreateSandboxCleanupRequestAwait(ctx context.Context, unitId int64, pollTime time.Duration) error {
	_, err := c.CreateSandboxCleanupRequest(ctx, unitId)
	if err != nil {
		return err
	}

	cleanupRequest, err := c.PollRequestFinished(ctx, unitId, pollTime, "cleanup")
	// After cleanup is finished it deletes itself and 404 is thrown
	if _, ok := err.(*ErrNotFound); ok {
		return nil
	}
	if err == nil && slices.Contains(cleanupRequest.Stages, "FAILED") {
		return fmt.Errorf("sandbox cleanup request finished with error")
	}
	return err
}

func (c *Client) CancelSandboxAllocationRequest(ctx context.Context, allocationRequestId int64) error {
	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/allocation-requests/%d/cancel", c.Endpoint, allocationRequestId), nil)
	if err != nil {
		return err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return err
	}

	if status == http.StatusNotFound {
		return &ErrNotFound{ResourceName: "sandbox allocation request", Identifier: strconv.FormatInt(allocationRequestId, 10)}
	}

	if status != http.StatusOK {
		return fmt.Errorf("status: %d, body: %s", status, body)
	}

	return nil
}

func (c *Client) GetSandboxRequestAnsibleOutputs(ctx context.Context, sandboxRequestId, page, pageSize int64, outputType string) (*SandboxRequestStageOutput, error) {
	query := url.Values{}
	query.Add("page", strconv.FormatInt(page, 10))
	query.Add("page_size", strconv.FormatInt(pageSize, 10))

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(
		"%s/kypo-sandbox-service/api/v1/allocation-requests/%d/stages/%s/outputs?%s", c.Endpoint, sandboxRequestId, outputType, query.Encode()), nil)
	if err != nil {
		return nil, err
	}

	body, status, err := c.doRequest(req)
	if err != nil {
		return nil, err
	}

	outputRaw := sandboxRequestStageOutputRaw{}

	if status == http.StatusNotFound {
		return nil, &ErrNotFound{ResourceName: "sandbox request", Identifier: strconv.FormatInt(sandboxRequestId, 10)}
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("status: %d, body: %s", status, body)
	}

	err = json.Unmarshal(body, &outputRaw)
	if err != nil {
		return nil, err
	}

	output := SandboxRequestStageOutput{
		Page:       outputRaw.Page,
		PageSize:   outputRaw.PageSize,
		PageCount:  outputRaw.PageCount,
		Count:      outputRaw.Count,
		TotalCount: outputRaw.TotalCount,
		Result:     "",
	}

	for _, line := range outputRaw.Results {
		output.Result += "\n" + line.Content
	}

	return &output, nil
}
