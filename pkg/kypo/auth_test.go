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
	"time"
)

func assertRequestToKeycloak(t *testing.T, request *http.Request) {
	assert.Equal(t, "application/x-www-form-urlencoded", request.Header.Get("Content-Type"))
	assert.Equal(t, "/keycloak/realms/KYPO/protocol/openid-connect/token", request.URL.Path)
	assert.Equal(t, "POST", request.Method)

	err := request.ParseForm()
	assert.NoError(t, err)

	assert.Equal(t, request.PostFormValue("username"), "username")
	assert.Equal(t, request.PostFormValue("password"), "password")
	assert.Equal(t, request.PostFormValue("client_id"), "client_id")
	assert.Equal(t, request.PostFormValue("grant_type"), "password")
}

func keycloakSuccessfulHandler(t *testing.T, writer http.ResponseWriter, request *http.Request) {
	assertRequestToKeycloak(t, request)

	r := struct {
		AccessToken      string `json:"access_token"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
		RefreshToken     string `json:"refresh_token"`
		TokenType        string `json:"token_type"`
		NotBeforePolicy  int    `json:"not-before-policy"`
		SessionState     string `json:"session_state"`
		Scope            string `json:"scope"`
	}{
		AccessToken:      "token",
		ExpiresIn:        60,
		RefreshExpiresIn: 30,
		RefreshToken:     "refresh_token",
		TokenType:        "Bearer",
		NotBeforePolicy:  0,
		SessionState:     "session_state",
		Scope:            "email openid profile",
	}
	response, _ := json.Marshal(r)
	_, _ = fmt.Fprint(writer, string(response))
}

func TestLoginKeycloakSuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		keycloakSuccessfulHandler(t, writer, request)
	}))
	defer ts.Close()

	c := kypo.Client{
		Endpoint:   ts.URL,
		ClientID:   "client_id",
		HTTPClient: http.DefaultClient,
		Token:      "old_token",
		Username:   "username",
		Password:   "password",
	}

	err := kypo.Authenticate(&c)

	assert.NoError(t, err)
	assert.Equal(t, ts.URL, c.Endpoint)
	assert.Equal(t, "client_id", c.ClientID)
	assert.Equal(t, http.DefaultClient, c.HTTPClient)
	assert.Equal(t, "token", c.Token)
	assert.WithinDuration(t, time.Now().Add(time.Duration(60)*time.Second), c.TokenExpiryTime, 100*time.Millisecond)
	assert.Equal(t, "username", c.Username)
	assert.Equal(t, "password", c.Password)
	assert.Equal(t, 0, c.RetryCount)
}

func TestLoginKeycloakUnsuccessful(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		assertRequestToKeycloak(t, request)

		r := struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}{
			Error:            "invalid_grant",
			ErrorDescription: "Invalid user credentials",
		}
		response, _ := json.Marshal(r)
		writer.WriteHeader(http.StatusUnauthorized)
		_, _ = fmt.Fprint(writer, string(response))
	}))
	defer ts.Close()

	c := kypo.Client{
		Endpoint:   ts.URL,
		ClientID:   "client_id",
		HTTPClient: http.DefaultClient,
		Token:      "token",
		Username:   "username",
		Password:   "password",
	}
	expected := fmt.Errorf("authentication to Keycloak failed, status: 401, body: " +
		"{\"error\":\"invalid_grant\",\"error_description\":\"Invalid user credentials\"}")

	err := kypo.Authenticate(&c)

	assert.Equal(t, expected, err)
	assert.Equal(t, ts.URL, c.Endpoint)
	assert.Equal(t, "client_id", c.ClientID)
	assert.Equal(t, http.DefaultClient, c.HTTPClient)
	assert.Equal(t, "token", c.Token)
	assert.Equal(t, "username", c.Username)
	assert.Equal(t, "password", c.Password)
	assert.Equal(t, 0, c.RetryCount)
}

func TestRefreshToken(t *testing.T) {
	requestCounter := 0
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requestCounter++
		keycloakSuccessfulHandler(t, writer, request)
	}))
	defer ts.Close()

	c := kypo.Client{
		Endpoint:   ts.URL,
		ClientID:   "client_id",
		HTTPClient: http.DefaultClient,
		Token:      "old_token",
		Username:   "username",
		Password:   "password",
	}

	err := kypo.RefreshToken(&c, context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, requestCounter)

	c.TokenExpiryTime = time.Now().Add(time.Hour)

	err = kypo.RefreshToken(&c, context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 0, requestCounter)

	c.TokenExpiryTime = time.Now()
	err = kypo.RefreshToken(&c, context.Background())
	assert.NoError(t, err)
	assert.Equal(t, 1, requestCounter)
	assert.Equal(t, "token", c.Token)
}
