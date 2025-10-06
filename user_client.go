package gowithings

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// UserClient represents a withings API client authenticated for accessing a specific users' data. The client will
// automatically handle updating the refresh token as needed.
type UserClient struct {
	clientID     string
	clientSecret string
	token        RequestToken
	httpClient   *http.Client
	sync.Mutex
}

// newRequest will create a new http request with a valid auth header. If the access token is
// expired it will automatically refresh it.
func (c *UserClient) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	c.Lock()
	defer c.Unlock()

	// Checking to determine if we need to try and refresh the token.
	switch {
	case c.token.AccessToken == "":
		return nil, fmt.Errorf("no access token provided")
	case time.Now().Sub(c.token.AccessTokenCreationDate).Hours() > 8760:
		return nil, fmt.Errorf("access token expired")
	case c.token.RefreshToken == "" || int(time.Now().Sub(c.token.RefreshTokenCreationDate).Seconds()) >= c.token.ExpiresIn:
		if err := c.refreshToken(ctx); err != nil {
			return nil, fmt.Errorf("failed to update refresh token: %w", err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %w", err)
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", c.token.AccessToken))

	return req, nil
}

// refreshToken updates the refresh token.
// Thread Safe: NO
func (c *UserClient) refreshToken(ctx context.Context) error {
	reqTokenResp := RequestTokenResponse{}

	req, err := http.NewRequest(http.MethodPost, RequestTokenURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %s", err)
	}

	reqQuery := req.URL.Query()
	reqQuery.Add("action", "requesttoken")
	reqQuery.Add("client_id", c.clientID)
	reqQuery.Add("client_secret", c.clientSecret)
	reqQuery.Add("grant_type", "refresh_token")
	reqQuery.Add("refresh_token", c.token.RefreshToken)
	req.URL.RawQuery = reqQuery.Encode()

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	createdAt := time.Now()
	resp, err := c.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to request token: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %s", err)
	}

	err = json.Unmarshal(body, &reqTokenResp)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response body: %s", err)
	}
	if reqTokenResp.Status != 0 {
		return fmt.Errorf("failed with status %d", reqTokenResp.Status)
	}
	reqTokenResp.Body.AccessTokenCreationDate = c.token.AccessTokenCreationDate
	reqTokenResp.Body.RefreshTokenCreationDate = createdAt

	c.token = reqTokenResp.Body
	return nil
}

// RefreshToken refreshes the access token.
// Thread Safe: YES
func (c *UserClient) RefreshToken(ctx context.Context) error {
	c.Lock()
	defer c.Unlock()

	return c.refreshToken(ctx)
}

// GetToken returns the current token for the user.
func (c *UserClient) GetToken() RequestToken {
	c.Lock()
	t := c.token
	c.Unlock()
	return t
}
