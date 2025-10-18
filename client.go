package gowithings

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

const (
	DefaultScopes    = "user.info,user.metrics,user.activity"
	AuthorizationURL = "https://account.withings.com/oauth2_user/authorize2"
	RequestTokenURL  = "https://wbsapi.withings.net/v2/oauth2"
	SignatureURL     = "https://wbsapi.withings.net/v2/signature"
)

// genStateValue generates a random 64 byte string that is URL encoded to be used
// as the state value when generating auth codes in Oauth2.
func genStateValue() (string, error) {
	buf := make([]byte, 64)
	_, err := rand.Read(buf)
	return base64.URLEncoding.EncodeToString(buf), err
}

// genHMACSHA256String generated a HMAC SHA256 for the message provided using the key provided.
func genHMACSHA256String(key, message string) string {
	// Create a new HMAC hash using SHA256 and the provided key.
	h := hmac.New(sha256.New, []byte(key))

	// Write the message bytes to the HMAC hash.
	h.Write([]byte(message))

	// Get the HMAC sum and encode it to a hexadecimal string.
	return hex.EncodeToString(h.Sum(nil))
}

// Config defines the configuration for a new client.
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// Client represents a client of the Withings API.
type Client struct {
	config     Config
	httpClient *http.Client
}

// NewClient creates a new client based on the configuration provided.
func NewClient(config Config) *Client {
	client := &Client{
		config:     config,
		httpClient: &http.Client{},
	}

	return client
}

// AuthCodeURL returns the authorization URL for users to authorize the client as well as a random state
// value to associate the response with.
func (client *Client) AuthCodeURL() (url string, state string, err error) {
	state, err = genStateValue()
	if err != nil {
		return "", "", err
	}
	url = fmt.Sprintf("%s?response_type=code&client_id=%s&scope=%s&redirect_uri=%s&state=%s",
		AuthorizationURL, client.config.ClientID, DefaultScopes, client.config.RedirectURL, state)

	return url, state, nil
}

// RequestToken request a token for the code provided.
func (client *Client) RequestToken(ctx context.Context, code string) (RequestTokenResponse, error) {
	reqTokenResp := RequestTokenResponse{}

	req, err := http.NewRequest(http.MethodPost, RequestTokenURL, nil)
	if err != nil {
		return reqTokenResp, fmt.Errorf("failed to create request: %s", err)
	}

	reqQuery := req.URL.Query()
	reqQuery.Add("action", "requesttoken")
	reqQuery.Add("client_id", client.config.ClientID)
	reqQuery.Add("client_secret", client.config.ClientSecret)
	reqQuery.Add("grant_type", "authorization_code")
	reqQuery.Add("redirect_uri", client.config.RedirectURL)
	reqQuery.Add("code", code)
	req.URL.RawQuery = reqQuery.Encode()

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	createdAt := time.Now()
	resp, err := client.httpClient.Do(req.WithContext(ctx))
	if err != nil {

		return reqTokenResp, fmt.Errorf("failed to request token: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return reqTokenResp, err
	}

	err = json.Unmarshal(body, &reqTokenResp)
	if err != nil {
		return reqTokenResp, err
	}
	reqTokenResp.Body.AccessTokenCreationDate = createdAt
	reqTokenResp.Body.RefreshTokenCreationDate = createdAt

	return reqTokenResp, nil
}

// NewUserClient creates a new UserClient for the user represented by the token provided.
func (client *Client) NewUserClient(token RequestToken) *UserClient {
	return &UserClient{
		clientID:     client.config.ClientID,
		clientSecret: client.config.ClientSecret,
		token:        token,
		httpClient:   &http.Client{},
	}
}

// NewUserClientFromRefreshToken generates a new user client from the refresh token provided.
// This is primarily used when loading a refresh token from a stored location to make
// a new request for that user.
func (client *Client) NewUserClientFromRefreshToken(ctx context.Context, refreshToken string, refreshTokenCreationDate time.Time) (*UserClient, error) {
	c := UserClient{
		clientID:     client.config.ClientID,
		clientSecret: client.config.ClientSecret,
		token: RequestToken{
			UserID:                   0,
			AccessToken:              "",
			RefreshToken:             refreshToken,
			ExpiresIn:                0,
			Scope:                    "",
			CSRFToken:                "",
			TokenType:                "",
			AccessTokenCreationDate:  time.Time{},
			RefreshTokenCreationDate: refreshTokenCreationDate,
		},
		httpClient: &http.Client{},
	}

	err := c.refreshToken(ctx)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

// DemoUser generates a UserClient for the demo user. This is primarily used for testing.
func (client *Client) DemoUser(ctx context.Context) (*UserClient, error) {
	reqTokenResp := RequestTokenResponse{}

	// Obtain nonce.
	nonce, err := client.getNonce(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %s", err)
	}

	signatureStr := genHMACSHA256String(client.config.ClientSecret, fmt.Sprintf("%s,%s,%s", "getdemoaccess", client.config.ClientID, nonce))

	req, err := http.NewRequest(http.MethodPost, RequestTokenURL, nil)

	if err != nil {
		return nil, fmt.Errorf("failed to create request: %s", err)
	}

	reqQuery := req.URL.Query()
	reqQuery.Add("action", "getdemoaccess")
	reqQuery.Add("client_id", client.config.ClientID)
	reqQuery.Add("nonce", nonce)
	reqQuery.Add("signature", signatureStr)
	reqQuery.Add("scope_oauth2", DefaultScopes)
	req.URL.RawQuery = reqQuery.Encode()

	createdAt := time.Now()
	resp, err := client.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, fmt.Errorf("failed to request token: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(body, &reqTokenResp)
	if err != nil {
		return nil, err
	}
	reqTokenResp.Body.AccessTokenCreationDate = createdAt
	reqTokenResp.Body.RefreshTokenCreationDate = createdAt

	uc := client.NewUserClient(reqTokenResp.Body)

	if reqTokenResp.Status != 0 {
		return nil, fmt.Errorf("failed with status %d", reqTokenResp.Status)
	}

	return uc, nil
}

type NonceRequestWrapper struct {
	Status int          `json:"status"`
	Body   NonceRequest `json:"body"`
}

type NonceRequest struct {
	Nonce string `json:"nonce"`
}

// getNonce retrieves a nonce from the withigns API.
func (client *Client) getNonce(ctx context.Context) (string, error) {
	req, err := http.NewRequest(http.MethodPost, SignatureURL, nil)

	ts := strconv.FormatInt(time.Now().Unix(), 10)
	signatureStr := genHMACSHA256String(client.config.ClientSecret, fmt.Sprintf("%s,%s,%s", "getnonce", client.config.ClientID, ts))

	reqQuery := req.URL.Query()
	reqQuery.Add("action", "getnonce")
	reqQuery.Add("client_id", client.config.ClientID)
	reqQuery.Add("timestamp", ts)
	reqQuery.Add("signature", signatureStr)
	req.URL.RawQuery = reqQuery.Encode()

	resp, err := client.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return "", fmt.Errorf("nonce request failed: %s", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	reqResp := NonceRequestWrapper{}
	err = json.Unmarshal(body, &reqResp)
	if err != nil {
		return "", err
	}
	if reqResp.Status != 0 {
		return "", fmt.Errorf("failed with status %d", reqResp.Status)
	}

	return reqResp.Body.Nonce, nil
}
