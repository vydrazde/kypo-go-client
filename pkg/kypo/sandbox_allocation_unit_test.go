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
	assert.Equal(t, http.MethodGet, request.Method)
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
	assert.Equal(t, http.MethodPost, request.Method)

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

func assertSandboxRequest(t *testing.T, request *http.Request, requestType string) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, fmt.Sprintf("/kypo-sandbox-service/api/v1/sandbox-allocation-units/1/%s-request", requestType), request.URL.Path)
	assert.Equal(t, http.MethodGet, request.Method)
}

func TestCreateSandboxAllocationUnitAwaitSuccessful(t *testing.T) {
	counter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		counter++
		if counter == 1 {
			assertSandboxAllocationUnitCreate(t, request)

			writer.WriteHeader(http.StatusCreated)
			response, _ := json.Marshal([]SandboxAllocationUnit{sandboxAllocationUnitResponse})
			_, _ = fmt.Fprint(writer, string(response))
			return
		}

		assertSandboxRequest(t, request, "allocation")
		r := SandboxAllocationRequest{
			Id:               1,
			AllocationUnitId: 1,
			Created:          "2023-11-26T17:04:20.032500+01:00",
			Stages:           []string{"FINISHED", "FINISHED", "FINISHED"},
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)
	expected := expectedSandboxAllocationUnit
	expected.AllocationRequest.Stages = []string{"FINISHED", "FINISHED", "FINISHED"}
	expected.AllocationRequest.Created = "2023-11-26T17:04:20.032500+01:00"

	actual, err := c.CreateSandboxAllocationUnitAwait(context.Background(), 1, 1)

	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)
	assert.Equal(t, 2, counter)
}

func TestCreateSandboxAllocationUnitAwaitSuccessfulWithDelay(t *testing.T) {
	counter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		counter++
		if counter == 1 {
			assertSandboxAllocationUnitCreate(t, request)

			writer.WriteHeader(http.StatusCreated)
			response, _ := json.Marshal([]SandboxAllocationUnit{sandboxAllocationUnitResponse})
			_, _ = fmt.Fprint(writer, string(response))
			return
		}

		assertSandboxRequest(t, request, "allocation")
		r := SandboxAllocationRequest{
			Id:               1,
			AllocationUnitId: 1,
			Created:          "2023-11-26T17:04:20.032500+01:00",
			Stages:           []string{"FINISHED", "FINISHED", "RUNNING"},
		}

		switch counter {
		case 2:
			r.Stages = []string{"FINISHED", "FINISHED", "IN_QUEUE"}
		case 3:
			r.Stages = []string{"FINISHED", "FINISHED", "RUNNING"}
		default:
			r.Stages = []string{"FINISHED", "FINISHED", "FINISHED"}
		}

		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)
	expected := expectedSandboxAllocationUnit
	expected.AllocationRequest.Stages = []string{"FINISHED", "FINISHED", "FINISHED"}
	expected.AllocationRequest.Created = "2023-11-26T17:04:20.032500+01:00"

	actual, err := c.CreateSandboxAllocationUnitAwait(context.Background(), 1, 1)

	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)
	assert.Equal(t, 4, counter)
}

func assertSandboxAllocationUnitCleanup(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "/kypo-sandbox-service/api/v1/sandbox-allocation-units/1/cleanup-request", request.URL.Path)
	assert.Equal(t, http.MethodPost, request.Method)
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

func TestCleanupSandboxAllocationUnitAwaitSuccessful(t *testing.T) {
	counter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		counter++
		if counter == 1 {
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
			return
		}

		assertSandboxRequest(t, request, "cleanup")
		r := SandboxAllocationRequest{
			Id:               1,
			AllocationUnitId: 1,
			Created:          "2023-11-26T17:04:20.032500+01:00",
			Stages:           []string{"FINISHED", "FINISHED", "FINISHED"},
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	err := c.CreateSandboxCleanupRequestAwait(context.Background(), 1, 1)

	assert.NoError(t, err)
	assert.Equal(t, 2, counter)
}

func TestCleanupSandboxAllocationUnitAwaitSuccessfulWithDelay(t *testing.T) {
	counter := 0
	r := SandboxAllocationRequest{
		Id:               1,
		AllocationUnitId: 1,
		Created:          "2023-11-26T17:04:20.032500+01:00",
		Stages:           []string{"IN_QUEUE", "IN_QUEUE", "IN_QUEUE"},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		counter++
		if counter == 1 {
			assertSandboxAllocationUnitCleanup(t, request)

			writer.WriteHeader(http.StatusCreated)
			response, _ := json.Marshal(r)
			_, _ = fmt.Fprint(writer, string(response))
			return
		}

		assertSandboxRequest(t, request, "cleanup")

		switch counter {
		case 2:
			r.Stages = []string{"FINISHED", "FINISHED", "IN_QUEUE"}
		case 3:
			r.Stages = []string{"FINISHED", "FINISHED", "RUNNING"}
		default:
			r.Stages = []string{"FINISHED", "FINISHED", "FINISHED"}
		}

		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	err := c.CreateSandboxCleanupRequestAwait(context.Background(), 1, 1)

	assert.NoError(t, err)
	assert.Equal(t, 4, counter)
}

func TestCleanupSandboxAllocationUnitAwaitFailed(t *testing.T) {
	counter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		counter++
		if counter == 1 {
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
			return
		}

		assertSandboxRequest(t, request, "cleanup")
		r := SandboxAllocationRequest{
			Id:               1,
			AllocationUnitId: 1,
			Created:          "2023-11-26T17:04:20.032500+01:00",
			Stages:           []string{"FINISHED", "FINISHED", "FAILED"},
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)
	expected := &kypo.Error{
		ResourceName: "sandbox cleanup request",
		Identifier:   "sandbox allocation unit 1",
		Err:          fmt.Errorf("sandbox cleanup request finished with error"),
	}

	err := c.CreateSandboxCleanupRequestAwait(context.Background(), 1, 1)

	assert.Equal(t, expected, err)
	assert.Equal(t, 2, counter)
}

func TestCancelSandboxAllocationRequestSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
		assert.Equal(t, "/kypo-sandbox-service/api/v1/allocation-requests/1/cancel", request.URL.Path)
		assert.Equal(t, http.MethodPatch, request.Method)
	}))
	defer ts.Close()

	c := minimalClient(ts)

	err := c.CancelSandboxAllocationRequest(context.Background(), 1)

	assert.NoError(t, err)
}

func TestGetSandboxAllocationRequestOutputSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
		assert.Equal(t, "/kypo-sandbox-service/api/v1/allocation-requests/1/stages/user-ansible/outputs", request.URL.Path)
		assert.Equal(t, http.MethodGet, request.Method)

		err := request.ParseForm()
		assert.NoError(t, err)
		assert.Equal(t, "1", request.Form.Get("page"))
		assert.Equal(t, "10", request.Form.Get("page_size"))

		type Content struct {
			Content string `json:"content"`
		}
		r := struct {
			Page       int       `json:"page"`
			PageSize   int       `json:"page_size"`
			PageCount  int       `json:"page_count"`
			Count      int       `json:"count"`
			TotalCount int       `json:"total_count"`
			Results    []Content `json:"results"`
		}{
			Page:       1,
			PageSize:   10,
			PageCount:  10,
			Count:      10,
			TotalCount: 100,
			Results: []Content{
				{Content: "1"},
				{Content: "2"},
				{Content: "3"},
				{Content: "4"},
				{Content: "5"},
				{Content: "6"},
				{Content: "7"},
				{Content: "8"},
				{Content: "9"},
				{Content: "10"},
			},
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)
	expected := kypo.SandboxRequestStageOutput{
		Page:       1,
		PageSize:   10,
		PageCount:  10,
		Count:      10,
		TotalCount: 100,
		Result:     "1\n2\n3\n4\n5\n6\n7\n8\n9\n10\n",
	}

	actual, err := c.GetSandboxRequestAnsibleOutputs(context.Background(), 1, 1, 10, "user-ansible")

	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)
}
