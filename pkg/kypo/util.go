package kypo

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

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
	body, err = io.ReadAll(res.Body)
	if err != nil {
		return
	}

	return
}

func (c *Client) doRequestWithRetry(req *http.Request, expectedStatusCode int, resourceName string, identifier any) (body []byte, statusCode int, err error) {
	duration := 50 * time.Millisecond
	timer := time.NewTimer(duration)
	defer timer.Stop()
	for i := 0; i <= c.RetryCount; i++ {
		body, statusCode, err = c.doRequest(req)
		if err != nil {
			return
		}
		switch statusCode {
		case expectedStatusCode:
			return
		case http.StatusNotFound:
			err = &Error{ResourceName: resourceName, Identifier: identifier, Err: ErrNotFound}
		default:
			err = &Error{ResourceName: resourceName, Identifier: identifier, Err: fmt.Errorf("status: %d, body: %s", statusCode, body)}
		}
		timer.Stop()
		duration *= 2
		timer.Reset(duration)
		select {
		case <-req.Context().Done():
			err = req.Context().Err()
			return
		case <-timer.C:
		}
	}
	// Only the last error will be returned
	// Aggregating the errors in a readable way seems overly complex
	return
}

func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}
