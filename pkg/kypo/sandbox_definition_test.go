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

type User struct {
	Id         int    `json:"id"`
	Sub        string `json:"sub"`
	FullName   string `json:"full_name"`
	GivenName  string `json:"given_name"`
	FamilyName string `json:"family_name"`
	Mail       string `json:"mail"`
}

var sandboxDefinitionResponse = struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Url       string `json:"url"`
	Rev       string `json:"rev"`
	CreatedBy User   `json:"created_by"`
}{
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
}

func assertSandboxDefinitionGet(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "/kypo-sandbox-service/api/v1/definitions/1", request.URL.Path)
	assert.Equal(t, "GET", request.Method)
}

func TestGetSandboxDefinitionSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxDefinitionGet(t, request)

		response, _ := json.Marshal(sandboxDefinitionResponse)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := kypo.SandboxDefinition{
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
	}

	actual, err := c.GetSandboxDefinition(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)
}

func TestGetSandboxDefinitionNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxDefinitionGet(t, request)

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
		ResourceName: "sandbox definition",
		Identifier:   int64(1),
		Err:          kypo.ErrNotFound,
	}

	td, actual := c.GetSandboxDefinition(context.Background(), 1)

	assert.Nil(t, td)
	assert.Equal(t, expected, actual)
}

func TestGetSandboxDefinitionServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxDefinitionGet(t, request)

		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := &kypo.Error{
		ResourceName: "sandbox definition",
		Identifier:   int64(1),
		Err:          fmt.Errorf("status: 500, body: "),
	}

	td, actual := c.GetSandboxDefinition(context.Background(), 1)

	assert.Nil(t, td)
	assert.Equal(t, expected, actual)
}

func assertSandboxDefinitionCreate(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "/kypo-sandbox-service/api/v1/definitions", request.URL.Path)
	assert.Equal(t, "POST", request.Method)
}

func TestCreateSandboxDefinitionSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxDefinitionCreate(t, request)

		writer.WriteHeader(http.StatusCreated)
		response, _ := json.Marshal(sandboxDefinitionResponse)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := kypo.SandboxDefinition{
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
	}

	actual, err := c.CreateSandboxDefinition(context.Background(), "url", "rev")

	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)
}

func TestCreateSandboxDefinitionServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxDefinitionCreate(t, request)

		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := &kypo.Error{
		ResourceName: "sandbox definition",
		Identifier:   "",
		Err:          fmt.Errorf("status: 500, body: "),
	}

	actual, err := c.CreateSandboxDefinition(context.Background(), "url", "rev")

	assert.Nil(t, actual)
	assert.Equal(t, expected, err)
}

func assertSandboxDefinitionDelete(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "/kypo-sandbox-service/api/v1/definitions/1", request.URL.Path)
	assert.Equal(t, "DELETE", request.Method)
}

func TestDeleteSandboxDefinitionSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxDefinitionDelete(t, request)

		writer.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	c := minimalClient(ts)

	err := c.DeleteSandboxDefinition(context.Background(), 1)

	assert.NoError(t, err)
}

func TestDeleteSandboxDefinitionNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxDefinitionDelete(t, request)

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
		ResourceName: "sandbox definition",
		Identifier:   int64(1),
		Err:          kypo.ErrNotFound,
	}

	actual := c.DeleteSandboxDefinition(context.Background(), 1)

	assert.Equal(t, expected, actual)
}

func TestDeleteSandboxDefinitionServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertSandboxDefinitionDelete(t, request)

		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := minimalClient(ts)
	expected := &kypo.Error{
		ResourceName: "sandbox definition",
		Identifier:   int64(1),
		Err:          fmt.Errorf("status: 500, body: "),
	}
	actual := c.DeleteSandboxDefinition(context.Background(), 1)

	assert.Equal(t, expected, actual)
}
