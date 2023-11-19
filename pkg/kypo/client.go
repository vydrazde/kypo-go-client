package kypo

import (
	"net/http"
	"time"
)

// Client struct stores information for authentication to the KYPO API.
// All functions are methods of this struct
type Client struct {
	Endpoint        string
	ClientID        string
	HTTPClient      *http.Client
	Token           string
	TokenExpiryTime time.Time
	Username        string
	Password        string
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
