package kypo

import (
	"net/http"
	"time"
)

// Client struct stores information for authentication to the KYPO API.
// All functions are methods of this struct
type Client struct {
	// Endpoint of the KYPO instance to connect to. For example `https://your.kypo.ex`.
	Endpoint string

	// ClientID used by the KYPO instance OIDC provider.
	ClientID string

	// HTTPClient which is used to do requests.
	HTTPClient *http.Client

	// Bearer Token which is used for authentication to the KYPO instance. Is set by NewClient function.
	Token string

	// Time when Token expires, used to refresh it automatically when required. Is set by NewClient function.
	// Is used only with KYPO instances using Keycloak OIDC provider.
	TokenExpiryTime time.Time

	// Username of the user to login as.
	Username string

	// Password of the user to login as.
	Password string

	// How many times should a failed HTTP request be retried. There is a delay of 100ms before the first retry.
	// The delay is doubled before each following retry.
	RetryCount int
}

// NewClientWithToken creates and returns a Client which uses an already created Bearer token.
func NewClientWithToken(endpoint, clientId, token string) (*Client, error) {
	client := Client{
		Endpoint:   endpoint,
		ClientID:   clientId,
		HTTPClient: http.DefaultClient,
		Token:      token,
	}

	return &client, nil
}

// NewClient creates and returns a Client which uses username and password for authentication.
// The username and password is used to login to Keycloak of the KYPO instance. If the login fails,
// login to the legacy CSIRT-MU dummy OIDC issuer is attempted.
func NewClient(endpoint, clientId, username, password string) (*Client, error) {
	client := Client{
		Endpoint:   endpoint,
		ClientID:   clientId,
		HTTPClient: http.DefaultClient,
		Username:   username,
		Password:   password,
	}
	err := client.authenticate()
	if err != nil {
		return nil, err
	}
	return &client, nil
}
