package kypo

import (
	"io/ioutil"
	"net/http"
)

type UserModel struct {
	Id         int64  `json:"id" tfsdk:"id"`
	Sub        string `json:"sub" tfsdk:"sub"`
	FullName   string `json:"full_name" tfsdk:"full_name"`
	GivenName  string `json:"given_name" tfsdk:"given_name"`
	FamilyName string `json:"family_name" tfsdk:"family_name"`
	Mail       string `json:"mail" tfsdk:"mail"`
}

func (c *Client) doRequest(req *http.Request) (body []byte, statusCode int, err error) {
	err = c.refreshToken(req.Context())
	if err != nil {
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.Token)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return
	}
	defer func() {
		err2 := res.Body.Close()
		// If there was an error already, I assume it is more important
		if err == nil {
			err = err2
		}
	}()
	statusCode = res.StatusCode
	body, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	return
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
