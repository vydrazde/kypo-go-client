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

var trainingDefinitionJsonString = `{"title":"title","description":"description","prerequisites":[],"outcomes":[],"state":"UNRELEASED","show_stepper_bar":true,"levels":[],"estimated_duration":0,"variant_sandboxes":false}`

func minimalClient(ts *httptest.Server) kypo.Client {
	c := kypo.Client{
		Endpoint:   ts.URL,
		HTTPClient: http.DefaultClient,
		Token:      "token",
	}
	return c
}

func assertTrainingDefinitionGet(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "application/octet-stream", request.Header.Get("accept"))
	assert.Equal(t, "/kypo-rest-training/api/v1/exports/training-definitions/1", request.URL.Path)
	assert.Equal(t, "GET", request.Method)
}

func TestGetTrainingDefinitionSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertTrainingDefinitionGet(t, request)

		r := struct {
			Title             string   `json:"title"`
			Description       string   `json:"description"`
			Prerequisites     []string `json:"prerequisites"`
			Outcomes          []string `json:"outcomes"`
			State             string   `json:"state"`
			ShowStepperBar    bool     `json:"show_stepper_bar"`
			Levels            []string `json:"levels"`
			EstimatedDuration int      `json:"estimated_duration"`
			VariantSandboxes  bool     `json:"variant_sandboxes"`
		}{
			Title:             "title",
			Description:       "description",
			Prerequisites:     []string{},
			Outcomes:          []string{},
			State:             "UNRELEASED",
			ShowStepperBar:    true,
			Levels:            []string{},
			EstimatedDuration: 0,
			VariantSandboxes:  false,
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := kypo.TrainingDefinition{
		Id:      1,
		Content: trainingDefinitionJsonString,
	}

	actual, err := c.GetTrainingDefinition(context.Background(), 1)

	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)
}

func TestGetTrainingDefinitionNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertTrainingDefinitionGet(t, request)

		writer.WriteHeader(http.StatusNotFound)
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := &kypo.Error{
		ResourceName: "training definition",
		Identifier:   int64(1),
		Err:          kypo.ErrNotFound,
	}

	td, actual := c.GetTrainingDefinition(context.Background(), 1)

	assert.Nil(t, td)
	assert.Equal(t, expected, actual)
}

func TestGetTrainingDefinitionServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertTrainingDefinitionGet(t, request)

		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := &kypo.Error{
		ResourceName: "training definition",
		Identifier:   int64(1),
		Err:          fmt.Errorf("status: 500, body: "),
	}

	td, actual := c.GetTrainingDefinition(context.Background(), 1)

	assert.Nil(t, td)
	assert.Equal(t, expected, actual)
}

func assertTrainingDefinitionCreate(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "/kypo-rest-training/api/v1/imports/training-definitions", request.URL.Path)
	assert.Equal(t, "POST", request.Method)
}

func TestCreateTrainingDefinitionSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertTrainingDefinitionCreate(t, request)

		r := struct {
			Id                 int      `json:"id"`
			Title              string   `json:"title"`
			Description        string   `json:"description"`
			Prerequisites      []string `json:"prerequisites"`
			Outcomes           []string `json:"outcomes"`
			State              string   `json:"state"`
			BetaTestingGroupId *string  `json:"beta_testing_group_id"`
			ShowStepperBar     bool     `json:"show_stepper_bar"`
			Levels             []string `json:"levels"`
			CanBeArchived      bool     `json:"can_be_archived"`
			EstimatedDuration  int      `json:"estimated_duration"`
			LastEdited         string   `json:"last_edited"`
			LastEditedBy       string   `json:"last_edited_by"`
		}{
			Id:                 1,
			Title:              "title",
			Description:        "description",
			Prerequisites:      []string{},
			Outcomes:           []string{},
			State:              "UNRELEASED",
			BetaTestingGroupId: nil,
			ShowStepperBar:     true,
			Levels:             []string{},
			CanBeArchived:      false,
			EstimatedDuration:  0,
			LastEdited:         "2023-11-26T09:03:35.313174494Z",
			LastEditedBy:       "User 1",
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := kypo.TrainingDefinition{
		Id:      1,
		Content: trainingDefinitionJsonString,
	}

	actual, err := c.CreateTrainingDefinition(context.Background(), trainingDefinitionJsonString)

	assert.NoError(t, err)
	assert.Equal(t, &expected, actual)
}

func TestCreateTrainingDefinitionServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertTrainingDefinitionCreate(t, request)

		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := minimalClient(ts)

	expected := &kypo.Error{
		ResourceName: "training definition",
		Identifier:   "",
		Err:          fmt.Errorf("status: 500, body: "),
	}

	td, actual := c.CreateTrainingDefinition(context.Background(), trainingDefinitionJsonString)

	assert.Nil(t, td)
	assert.Equal(t, expected, actual)
}

func assertTrainingDefinitionDelete(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/json", request.Header.Get("Content-Type"))
	assert.Equal(t, "Bearer token", request.Header.Get("Authorization"))
	assert.Equal(t, "/kypo-rest-training/api/v1/training-definitions/1", request.URL.Path)
	assert.Equal(t, "DELETE", request.Method)
}

func TestDeleteTrainingDefinitionSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertTrainingDefinitionDelete(t, request)

		writer.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := minimalClient(ts)

	err := c.DeleteTrainingDefinition(context.Background(), 1)

	assert.NoError(t, err)
}

func TestDeleteTrainingDefinitionNotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertTrainingDefinitionDelete(t, request)

		writer.WriteHeader(http.StatusNotFound)
		type EntityErrorDetail struct {
			Entity          string `json:"entity"`
			Identifier      string `json:"identifier"`
			IdentifierValue int    `json:"identifier_value"`
			Reason          string `json:"reason"`
		}

		r := struct {
			Timestamp         int               `json:"timestamp"`
			Status            string            `json:"status"`
			Message           string            `json:"message"`
			Errors            []*string         `json:"errors"`
			Path              string            `json:"path"`
			EntityErrorDetail EntityErrorDetail `json:"entity_error_detail"`
		}{
			Timestamp: 1700990353304,
			Status:    "NOT_FOUND",
			Message:   "Entity TrainingDefinition (id: 1) not found.",
			Errors:    []*string{nil},
			Path:      "/kypo-rest-training/api/v1/training-definitions/1",
			EntityErrorDetail: EntityErrorDetail{
				Entity:          "TrainingDefinition",
				Identifier:      "id",
				IdentifierValue: 1,
				Reason:          "Entity TrainingDefinition (id: 1) not found.",
			},
		}
		response, _ := json.Marshal(r)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := minimalClient(ts)
	expected := &kypo.Error{
		ResourceName: "training definition",
		Identifier:   int64(1),
		Err:          kypo.ErrNotFound,
	}

	actual := c.DeleteTrainingDefinition(context.Background(), 1)

	assert.Equal(t, expected, actual)
}

func TestDeleteTrainingDefinitionServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertTrainingDefinitionDelete(t, request)

		writer.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := minimalClient(ts)
	expected := &kypo.Error{
		ResourceName: "training definition",
		Identifier:   int64(1),
		Err:          fmt.Errorf("status: 500, body: "),
	}
	actual := c.DeleteTrainingDefinition(context.Background(), 1)

	assert.Equal(t, expected, actual)
}
