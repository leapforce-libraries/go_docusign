package docusign

import (
	errortools "github.com/leapforce-libraries/go_errortools"
	oauth2 "github.com/leapforce-libraries/go_oauth2"
)

// Token stures Token object
//
type AccessToken struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int    `json:"expiresIn"`
}

func (service *Service) GetAccessToken() (*oauth2.Token, *errortools.Error) {
	/*body := struct {
		ClientID string `json:"clientId"`
		Secret   string `json:"secret"`
	}{
		service.clientID,
		service.clientSecret,
	}

	accessToken := AccessToken{}

	requestConfig := go_http.RequestConfig{
		URL:           AccessTokenURL,
		BodyModel:     body,
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
		ExpiresIn:   &expiresInJSON,
	}

	return &token, nil*/

	return nil, nil
}
