package docusign

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	jose "github.com/dvsekhvalnov/jose2go"
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

type JWTHeader struct {
	Algorithm string `json:"alg"`
	Type      string `json:"typ"`
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
	urlDomain := TokenURL
	if service.isDemo {
		urlDomain = TokenURLDemo
	}

	// jwt header
	jwtHeader := JWTHeader{"HS256", "JWT"}
	bHeader, err := json.Marshal(jwtHeader)
	if err != nil {
		return nil, errortools.ErrorMessage(err)
	}
	header := base64.URLEncoding.EncodeToString(bHeader)

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
	body := base64.URLEncoding.EncodeToString(bBody)

	fmt.Println(jwtBody)

	jwt, err := jose.Sign(fmt.Sprintf("%s.%s", header, body), jose.HS256, []byte(service.privateKey))

	fmt.Println(jwt)

	accessToken := AccessToken{}

	values := url.Values{}
	values.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	values.Set("assertion", jwt)

	requestConfig := go_http.RequestConfig{
		URL:           fmt.Sprintf("%s?%s", urlDomain, values.Encode()),
		ResponseModel: &accessToken,
	}

	_, _, e := service.httpRequest(http.MethodPost, &requestConfig, true)
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
