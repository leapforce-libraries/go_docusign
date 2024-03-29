package docusign

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	jose "github.com/dvsekhvalnov/jose2go"
	jose_rsa "github.com/dvsekhvalnov/jose2go/keys/rsa"
	errortools "github.com/leapforce-libraries/go_errortools"
	go_http "github.com/leapforce-libraries/go_http"
	oauth2 "github.com/leapforce-libraries/go_oauth2"
)

// Token stures Token object
//
type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type JWTBody struct {
	Issuer         string `json:"iss"`
	Subject        string `json:"sub"`
	ExpirationTime int64  `json:"exp"`
	IssuedAt       int64  `json:"iat"`
	Audience       string `json:"aud"`
	Scope          string `json:"scope"`
}

func (service *Service) GetAccessToken() (*oauth2.Token, *errortools.Error) {
	urlDomain := tokenURL
	if service.isDemo {
		urlDomain = tokenURLDemo
	}

	// jwt body
	now := time.Now()

	audience := strings.ReplaceAll(strings.ReplaceAll(urlDomain, "https://", ""), "/oauth/token", "")

	jwtBody := JWTBody{
		Issuer:         service.integrationKey,
		Subject:        service.userName,
		ExpirationTime: now.Add(time.Hour).Unix(),
		IssuedAt:       now.Unix(),
		Audience:       audience,
		Scope:          service.scopes,
	}
	bBody, err := json.Marshal(jwtBody)
	if err != nil {
		return nil, errortools.ErrorMessage(err)
	}

	// jwt signature
	privateKey, err := jose_rsa.ReadPrivate([]byte(service.privateKey))
	if err != nil {
		fmt.Println(err)
	}

	err = privateKey.Validate()
	if err != nil {
		fmt.Println(err)
	}

	jwt, err := jose.SignBytes(bBody, jose.RS256, privateKey, jose.Header("typ", "JWT"))
	if err != nil {
		fmt.Println(err)
	}

	accessToken := AccessToken{}

	values := url.Values{}
	values.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	values.Set("assertion", jwt)

	requestConfig := go_http.RequestConfig{
		Method:        http.MethodPost,
		URL:           fmt.Sprintf("%s?%s", urlDomain, values.Encode()),
		ResponseModel: &accessToken,
	}

	_, _, e := service.httpRequestWithoutAccessToken(&requestConfig)
	if e != nil {
		return nil, e
	}

	expiresIn, _ := json.Marshal(accessToken.ExpiresIn)
	expiresInJSON := json.RawMessage(expiresIn)

	token := oauth2.Token{
		AccessToken: &accessToken.AccessToken,
		TokenType:   &accessToken.TokenType,
		ExpiresIn:   &expiresInJSON,
	}

	return &token, nil
}
