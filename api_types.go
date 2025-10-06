package gowithings

import "time"

type RequestTokenResponse struct {
	Status int          `json:"status"`
	Body   RequestToken `json:"body"`
}

type RequestToken struct {
	UserID                   int       `json:"user_id"`
	AccessToken              string    `json:"access_token"`
	RefreshToken             string    `json:"refresh_token"`
	ExpiresIn                int       `json:"expires_in"`
	Scope                    string    `json:"scope"`
	CSRFToken                string    `json:"csrf_token"`
	TokenType                string    `json:"token_type"`
	AccessTokenCreationDate  time.Time `json:"access_token_creation_date"`
	RefreshTokenCreationDate time.Time `json:"refresh_token_creation_date"`
}
