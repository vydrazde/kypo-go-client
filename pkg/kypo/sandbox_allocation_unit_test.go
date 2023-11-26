package kypo_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vydrazde/kypo-go-client/pkg/kypo"
	"net/http"
	"net/http/httptest"
	"testing"
)

type SandboxAllocationRequest struct {
	Id               int      `json:"id"`
	AllocationUnitId int      `json:"allocation_unit_id"`
	Created          string   `json:"created"`
	Stages           []string `json:"stages"`
}

type SandboxAllocationUnit struct {
	Id                int                       `json:"id"`
	PoolId            int                       `json:"pool_id"`
	AllocationRequest *SandboxAllocationRequest `json:"allocation_request"`
	CleanupRequest    *SandboxAllocationRequest `json:"cleanup_request"`
	CreatedBy         User                      `json:"created_by"`
	Locked            bool                      `json:"locked"`
}

var (
	sandboxAllocationUnitResponse = SandboxAllocationUnit{
		Id:     1,
		PoolId: 1,
		AllocationRequest: &SandboxAllocationRequest{
			Id:               1,
			AllocationUnitId: 1,
			Created:          "2023-10-23T11:58:21.757093+02:00",
			Stages:           []string{"FINISHED", "FINISHED", "FINISHED"},
		},
		CleanupRequest: nil,
		CreatedBy: User{
			Id:         1,
			Sub:        "sub",
			FullName:   "full_name",
			GivenName:  "give_name",
			FamilyName: "family_name",
			Mail:       "mail",
		},
		Locked: false,
	}
	expectedSandboxAllocationUnit = kypo.SandboxAllocationUnit{
		Id:     1,
		PoolId: 1,
		AllocationRequest: kypo.SandboxRequest{
			Id:               1,
			AllocationUnitId: 1,
			Created:          "2023-10-23T11:58:21.757093+02:00",
			Stages:           []string{"FINISHED", "FINISHED", "FINISHED"},
		},
		CleanupRequest: kypo.SandboxRequest{},
		CreatedBy: kypo.User{
			Id:         1,
			Sub:        "sub",
			FullName:   "full_name",
			GivenName:  "give_name",
			FamilyName: "family_name",
			Mail:       "mail",
		},
		Locked: false,
	}
)

func assertSandboxAllocationUnitGet(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "/kypo-sandbox-service/api/v1/sandbox-allocation-units/1", request.URL.Path)
	assert.Equal(t, "GET", request.Method)
}

func TestGetSandboxAllocationUnitSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxAllocationUnitGet(t, request)

		response, _ := json.Marshal(sandboxAllocationUnitResponse)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	actual, err := c.GetSandboxAllocationUnit(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, &expectedSandboxAllocationUnit, actual)
}

func TestGetSandboxAllocationUnitNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxAllocationUnitGet(t, request)

		writer.WriteHeader(http.StatusNotFound)
		r := struct {
			Detail string `json:"detail"`
		}{
			Detail: "No SandboxAllocationUnit matches the given query",
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := &kypo.Error{
		ResourceName: "sandbox allocation unit",
		Identifier:   int64(1),
		Err:          kypo.ErrNotFound,
	}

	td, actual := c.GetSandboxAllocationUnit(context.Background(), 1)

	assert.Nil(t, td)
	assert.Equal(t, expected, actual)
}

func TestGetSandboxAllocationUnitServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxAllocationUnitGet(t, request)

		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := &kypo.Error{
		ResourceName: "sandbox allocation unit",
		Identifier:   int64(1),
		Err:          fmt.Errorf("status: 500, body: "),
	}

	td, actual := c.GetSandboxAllocationUnit(context.Background(), 1)

	assert.Nil(t, td)
	assert.Equal(t, expected, actual)
}

func assertSandboxAllocationUnitCreate(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "/kypo-sandbox-service/api/v1/pools/1/sandbox-allocation-units", request.URL.Path)
	assert.Equal(t, "POST", request.Method)

	err := request.ParseForm()
	assert.NoError(t, err)
	assert.Equal(t, "1", request.Form.Get("count"))
}

func TestCreateSandboxAllocationUnitSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxAllocationUnitCreate(t, request)

		writer.WriteHeader(http.StatusCreated)
		response, _ := json.Marshal([]SandboxAllocationUnit{sandboxAllocationUnitResponse})
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := []kypo.SandboxAllocationUnit{expectedSandboxAllocationUnit}

	actual, err := c.CreateSandboxAllocationUnits(context.Background(), 1, 1)

	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestCreateSandboxAllocationUnitNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxAllocationUnitCreate(t, request)

		writer.WriteHeader(http.StatusNotFound)
		r := struct {
			Detail string `json:"detail"`
		}{
			Detail: "No Pool matches the given query",
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := &kypo.Error{
		ResourceName: "sandbox allocation units",
		Identifier:   "sandbox pool 1",
		Err:          kypo.ErrNotFound,
	}

	actual, err := c.CreateSandboxAllocationUnits(context.Background(), 1, 1)

	assert.Nil(t, actual)
	assert.Equal(t, expected, err)
}

func TestCreateSandboxAllocationUnitServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxAllocationUnitCreate(t, request)

		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := &kypo.Error{
		ResourceName: "sandbox allocation units",
		Identifier:   "sandbox pool 1",
		Err:          fmt.Errorf("status: 500, body: "),
	}

	actual, err := c.CreateSandboxAllocationUnits(context.Background(), 1, 1)

	assert.Nil(t, actual)
	assert.Equal(t, expected, err)
}

func assertSandboxAllocationUnitCleanup(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "/kypo-sandbox-service/api/v1/sandbox-allocation-units/1/cleanup-request", request.URL.Path)
	assert.Equal(t, "POST", request.Method)
}

func TestCleanupSandboxAllocationUnitSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxAllocationUnitCleanup(t, request)

		writer.WriteHeader(http.StatusCreated)
		r := SandboxAllocationRequest{
			Id:               1,
			AllocationUnitId: 1,
			Created:          "2023-11-26T17:04:20.032500+01:00",
			Stages:           []string{"IN_QUEUE", "IN_QUEUE", "IN_QUEUE"},
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	actual, err := c.CreateSandboxCleanupRequest(context.Background(), 1)
	expected := kypo.SandboxRequest{
		Id:               1,
		AllocationUnitId: 1,
		Created:          "2023-11-26T17:04:20.032500+01:00",
		Stages:           []string{"IN_QUEUE", "IN_QUEUE", "IN_QUEUE"},
	}

	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)
}

func TestCleanupSandboxAllocationUnitNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxAllocationUnitCleanup(t, request)

		writer.WriteHeader(http.StatusNotFound)
		r := struct {
			Detail string `json:"detail"`
		}{
			Detail: "No SandboxAllocationUnit matches the given query",
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)
	expected := &kypo.Error{
		ResourceName: "sandbox cleanup request",
		Identifier:   "sandbox allocation unit 1",
		Err:          kypo.ErrNotFound,
	}

	actual, err := c.CreateSandboxCleanupRequest(context.Background(), 1)

	assert.Nil(t, actual)
	assert.Equal(t, expected, err)
}

func TestCleanupSandboxAllocationUnitServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxAllocationUnitCleanup(t, request)

		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := minimalClient(ts)
	expected := &kypo.Error{
		ResourceName: "sandbox cleanup request",
		Identifier:   "sandbox allocation unit 1",
		Err:          fmt.Errorf("status: 500, body: "),
	}
	actual, err := c.CreateSandboxCleanupRequest(context.Background(), 1)

	assert.Nil(t, actual)
	assert.Equal(t, expected, err)
}
