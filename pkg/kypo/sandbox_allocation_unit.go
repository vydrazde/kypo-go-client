package kypo

import (
	"context"
	"encoding/json"
	"errors"
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
	CreatedBy         User           `json:"created_by" tfsdk:"created_by"`
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

// GetSandboxAllocationUnit reads a sandbox allocation unit based on its id.
func (c *Client) GetSandboxAllocationUnit(ctx context.Context, unitId int64) (*SandboxAllocationUnit, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/sandbox-allocation-units/%d", c.Endpoint, unitId), nil)
	if err != nil {
		return nil, err
	}

	body, _, err := c.doRequestWithRetry(req, http.StatusOK, "sandbox allocation unit", unitId)
	if err != nil {
		return nil, err
	}

	allocationUnit := SandboxAllocationUnit{}

	err = json.Unmarshal(body, &allocationUnit)
	if err != nil {
		return nil, err
	}

	return &allocationUnit, nil
}

// CreateSandboxAllocationUnits starts the allocation of `count` sandboxes in the sandbox pool specified by `poolId`.
func (c *Client) CreateSandboxAllocationUnits(ctx context.Context, poolId, count int64) ([]SandboxAllocationUnit, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/pools/%d/sandbox-allocation-units?count=%d", c.Endpoint, poolId, count), nil)
	if err != nil {
		return nil, err
	}

	body, _, err := c.doRequestWithRetry(req, http.StatusCreated, "sandbox allocation units", fmt.Sprintf("sandbox pool %d", poolId))
	if err != nil {
		return nil, err
	}

	var allocationUnit []SandboxAllocationUnit
	err = json.Unmarshal(body, &allocationUnit)
	if err != nil {
		return nil, err
	}

	return allocationUnit, nil
}

// CreateSandboxAllocationUnitAwait creates a single sandbox allocation unit and waits until its allocation finishes.
// Once the allocation is started, the status is checked once every `pollTime` elapses.
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

// CreateSandboxCleanupRequest starts a cleanup request for the specified sandbox allocation unit.
func (c *Client) CreateSandboxCleanupRequest(ctx context.Context, unitId int64) (*SandboxRequest, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/sandbox-allocation-units/%d/cleanup-request", c.Endpoint, unitId), nil)
	if err != nil {
		return nil, err
	}

	body, _, err := c.doRequestWithRetry(req, http.StatusCreated, "sandbox cleanup request", fmt.Sprintf("sandbox allocation unit %d", unitId))
	if err != nil {
		return nil, err
	}

	sandboxRequest := SandboxRequest{}
	err = json.Unmarshal(body, &sandboxRequest)
	if err != nil {
		return nil, err
	}

	return &sandboxRequest, nil
}

// PollRequestFinished periodically checks whether the specified request on given allocation unit has finished.
// The `requestType` should be one of `allocation` or `cleanup`. The check is done once every `pollTime` elapses.
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
				return nil, &Error{ResourceName: "sandbox request", Identifier: unitId, Err: ErrNotFound}
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

// CreateSandboxCleanupRequestAwait starts the cleanup request for the given sandbox allocation unit and waits until it finishes.
// Once the cleanup is started, the status is checked once every `pollTime` elapses.
func (c *Client) CreateSandboxCleanupRequestAwait(ctx context.Context, unitId int64, pollTime time.Duration) error {
	_, err := c.CreateSandboxCleanupRequest(ctx, unitId)
	if err != nil {
		return err
	}

	cleanupRequest, err := c.PollRequestFinished(ctx, unitId, pollTime, "cleanup")
	// After cleanup is finished it deletes itself and 404 is thrown
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err == nil && slices.Contains(cleanupRequest.Stages, "FAILED") {
		return &Error{ResourceName: "sandbox cleanup request", Identifier: fmt.Sprintf("sandbox allocation unit %d", unitId),
			Err: fmt.Errorf("sandbox cleanup request finished with error")}
	}
	return err
}

// CancelSandboxAllocationRequest sends a request to cancel the given allocation request.
func (c *Client) CancelSandboxAllocationRequest(ctx context.Context, allocationRequestId int64) error {
	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/kypo-sandbox-service/api/v1/allocation-requests/%d/cancel", c.Endpoint, allocationRequestId), nil)
	if err != nil {
		return err
	}

	_, _, err = c.doRequestWithRetry(req, http.StatusOK, "sandbox allocation request", allocationRequestId)
	if err != nil {
		return err
	}

	return nil
}

// GetSandboxRequestAnsibleOutputs reads the output of given allocation request stage.
// The `outputType` should be one of `user-ansible`, `networking-ansible` or `terraform`.
func (c *Client) GetSandboxRequestAnsibleOutputs(ctx context.Context, sandboxRequestId, page, pageSize int64, outputType string) (*SandboxRequestStageOutput, error) {
	query := url.Values{}
	query.Add("page", strconv.FormatInt(page, 10))
	query.Add("page_size", strconv.FormatInt(pageSize, 10))

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf(
		"%s/kypo-sandbox-service/api/v1/allocation-requests/%d/stages/%s/outputs?%s", c.Endpoint, sandboxRequestId, outputType, query.Encode()), nil)
	if err != nil {
		return nil, err
	}

	body, _, err := c.doRequestWithRetry(req, http.StatusOK, "sandbox request output", sandboxRequestId)
	if err != nil {
		return nil, err
	}

	outputRaw := sandboxRequestStageOutputRaw{}

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
		output.Result += line.Content + "\n"
	}

	return &output, nil
}
