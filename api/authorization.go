package api

import (
	"net/url"
)

const authorizeURL = "https://raindrop.io/oauth/authorize"

func CreateAuthorizationURL(clientID, redirectURI string) string {
	values := url.Values{}
	values.Set("response_type", "code")
	values.Set("client_id", clientID)
	values.Set("redirect_uri", redirectURI)
	return authorizeURL + "?" + values.Encode()
}
