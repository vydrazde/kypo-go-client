package kypo_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/vydrazde/kypo-go-client/pkg/kypo"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

type SandboxPool struct {
	Id            int               `json:"id"`
	Size          int               `json:"size"`
	MaxSize       int               `json:"max_size"`
	LockId        *int              `json:"lock_id"`
	Rev           string            `json:"rev"`
	RevSha        string            `json:"rev_sha"`
	CreatedBy     User              `json:"created_by"`
	HardwareUsage HardwareUsage     `json:"hardware_usage"`
	Definition    SandboxDefinition `json:"definition"`
}

type HardwareUsage struct {
	Vcpu      string `json:"vcpu"`
	Ram       string `json:"ram"`
	Instances string `json:"instances"`
	Network   string `json:"network"`
	Subnet    string `json:"subnet"`
	Port      string `json:"port"`
}

var (
	sandboxPoolResponse = SandboxPool{
		Id:      1,
		Size:    0,
		MaxSize: 1,
		LockId:  nil,
		Rev:     "rev",
		RevSha:  "rev_sha",
		CreatedBy: User{
			Id:         1,
			Sub:        "sub",
			FullName:   "full_name",
			GivenName:  "given_name",
			FamilyName: "family_name",
			Mail:       "mail",
		},
		HardwareUsage: HardwareUsage{
			Vcpu:      "0.000",
			Ram:       "0.000",
			Instances: "0.000",
			Network:   "0.000",
			Subnet:    "0.000",
			Port:      "0.000",
		},
		Definition: SandboxDefinition{
			Id:   1,
			Name: "name",
			Url:  "url",
			Rev:  "rev",
			CreatedBy: User{
				Id:         1,
				Sub:        "sub",
				FullName:   "full_name",
				GivenName:  "given_name",
				FamilyName: "family_name",
				Mail:       "mail",
			},
		},
	}
	expectedPoolResponse = kypo.SandboxPool{
		Id:      1,
		Size:    0,
		MaxSize: 1,
		LockId:  0,
		Rev:     "rev",
		RevSha:  "rev_sha",
		CreatedBy: kypo.User{
			Id:         1,
			Sub:        "sub",
			FullName:   "full_name",
			GivenName:  "given_name",
			FamilyName: "family_name",
			Mail:       "mail",
		},
		HardwareUsage: kypo.HardwareUsage{
			Vcpu:      "0.000",
			Ram:       "0.000",
			Instances: "0.000",
			Network:   "0.000",
			Subnet:    "0.000",
			Port:      "0.000",
		},
		Definition: kypo.SandboxDefinition{
			Id:   1,
			Name: "name",
			Url:  "url",
			Rev:  "rev",
			CreatedBy: kypo.User{
				Id:         1,
				Sub:        "sub",
				FullName:   "full_name",
				GivenName:  "given_name",
				FamilyName: "family_name",
				Mail:       "mail",
			},
		},
	}
)

func assertSandboxPoolGet(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "/kypo-sandbox-service/api/v1/pools/1", request.URL.Path)
	assert.Equal(t, http.MethodGet, request.Method)
}

func TestGetSandboxPoolSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxPoolGet(t, request)

		response, _ := json.Marshal(sandboxPoolResponse)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	actual, err := c.GetSandboxPool(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, &expectedPoolResponse, actual)
}

func TestGetSandboxPoolNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxPoolGet(t, request)

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
		ResourceName: "sandbox pool",
		Identifier:   int64(1),
		Err:          kypo.ErrNotFound,
	}

	td, actual := c.GetSandboxPool(context.Background(), 1)

	assert.Nil(t, td)
	assert.Equal(t, expected, actual)
}

func TestGetSandboxPoolServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxPoolGet(t, request)

		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := &kypo.Error{
		ResourceName: "sandbox pool",
		Identifier:   int64(1),
		Err:          fmt.Errorf("status: 500, body: "),
	}

	td, actual := c.GetSandboxPool(context.Background(), 1)

	assert.Nil(t, td)
	assert.Equal(t, expected, actual)
}

func assertSandboxPoolCreate(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "/kypo-sandbox-service/api/v1/pools", request.URL.Path)
	assert.Equal(t, http.MethodPost, request.Method)

	body, err := io.ReadAll(request.Body)
	assert.NoError(t, err)

	parsedBody := struct {
		DefinitionId int `json:"definition_id"`
		MaxSize      int `json:"max_size"`
	}{}
	err = json.Unmarshal(body, &parsedBody)
	assert.NoError(t, err)

	assert.Equal(t, 1, parsedBody.DefinitionId)
	assert.Equal(t, 1, parsedBody.MaxSize)
}

func TestCreateSandboxPoolSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxPoolCreate(t, request)

		writer.WriteHeader(http.StatusCreated)
		response, _ := json.Marshal(sandboxPoolResponse)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	actual, err := c.CreateSandboxPool(context.Background(), 1, 1)

	assert.NoError(t, err)
	assert.Equal(t, &expectedPoolResponse, actual)
}

func TestCreateSandboxPoolNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxPoolCreate(t, request)

		writer.WriteHeader(http.StatusNotFound)
		r := struct {
			Detail string `json:"detail"`
		}{
			Detail: "No Definition matches the given query",
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := &kypo.Error{
		ResourceName: "sandbox pool",
		Identifier:   "sandbox definition 1",
		Err:          kypo.ErrNotFound,
	}

	actual, err := c.CreateSandboxPool(context.Background(), 1, 1)

	assert.Nil(t, actual)
	assert.Equal(t, expected, err)
}

func TestCreateSandboxPoolServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxPoolCreate(t, request)

		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := &kypo.Error{
		ResourceName: "sandbox pool",
		Identifier:   "sandbox definition 1",
		Err:          fmt.Errorf("status: 500, body: "),
	}

	actual, err := c.CreateSandboxPool(context.Background(), 1, 1)

	assert.Nil(t, actual)
	assert.Equal(t, expected, err)
}

func assertSandboxPoolDelete(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "/kypo-sandbox-service/api/v1/pools/1", request.URL.Path)
	assert.Equal(t, http.MethodDelete, request.Method)
}

func TestDeleteSandboxPoolSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxPoolDelete(t, request)

		writer.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := minimalClient(ts)

	err := c.DeleteSandboxPool(context.Background(), 1)

	assert.NoError(t, err)
}

func TestDeleteSandboxPoolNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxPoolDelete(t, request)

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
		ResourceName: "sandbox pool",
		Identifier:   int64(1),
		Err:          kypo.ErrNotFound,
	}

	actual := c.DeleteSandboxPool(context.Background(), 1)

	assert.Equal(t, expected, actual)
}

func TestDeleteSandboxPoolServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxPoolDelete(t, request)

		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := minimalClient(ts)
	expected := &kypo.Error{
		ResourceName: "sandbox pool",
		Identifier:   int64(1),
		Err:          fmt.Errorf("status: 500, body: "),
	}
	actual := c.DeleteSandboxPool(context.Background(), 1)

	assert.Equal(t, expected, actual)
}

func assertSandboxPoolCleanup(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "/kypo-sandbox-service/api/v1/pools/1/cleanup-requests", request.URL.Path)
	assert.Equal(t, http.MethodPost, request.Method)

	err := request.ParseForm()
	assert.NoError(t, err)
	assert.Equal(t, "false", request.Form.Get("force"))
}

func TestCleanupSandboxPoolSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxPoolCleanup(t, request)

		writer.WriteHeader(http.StatusAccepted)
		r := []SandboxAllocationRequest{
			{
				Id:               1,
				AllocationUnitId: 1,
				Created:          "2023-11-26T15:48:41.614625+01:00",
				Stages:           []string{"IN_QUEUE", "IN_QUEUE", "IN_QUEUE"},
			},
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	err := c.CleanupSandboxPool(context.Background(), 1, false)

	assert.NoError(t, err)
}

func TestCleanupSandboxPoolNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxPoolCleanup(t, request)

		writer.WriteHeader(http.StatusNotFound)
		r := struct {
			Detail string `json:"detail"`
		}{
			Detail: "The instance of Pool with {'pk': 1} not found.",
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)
	expected := &kypo.Error{
		ResourceName: "sandbox pool",
		Identifier:   int64(1),
		Err:          kypo.ErrNotFound,
	}

	actual := c.CleanupSandboxPool(context.Background(), 1, false)

	assert.Equal(t, expected, actual)
}

func TestCleanupSandboxPoolServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxPoolCleanup(t, request)

		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := minimalClient(ts)
	expected := &kypo.Error{
		ResourceName: "sandbox pool",
		Identifier:   int64(1),
		Err:          fmt.Errorf("status: 500, body: "),
	}
	actual := c.CleanupSandboxPool(context.Background(), 1, false)

	assert.Equal(t, expected, actual)
}
