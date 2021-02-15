package docusign

import (
	errortools "github.com/leapforce-libraries/go_errortools"
	go_http "github.com/leapforce-libraries/go_http"
)

// Token stures Token object
//
type UserInfo struct {
	UserID     string    `json:"sub"`
	Name       string    `json:"name"`
	GivenName  string    `json:"given_name"`
	FamilyName string    `json:"family_name"`
	Created    string    `json:"created"`
	Email      string    `json:"email"`
	Accounts   []Account `json:"accounts"`
}

type Account struct {
	AccountID      string `json:"account_id"`
	IsDefault      bool   `json:"is_default"`
	AccountName    string `json:"account_name"`
	BaseURI        string `json:"base_uri"`
	OrganizationID string `json:"organization_id"`
}

func (service *Service) GetUserInfo() (*UserInfo, *errortools.Error) {
	url := UserInfoURL
	if service.isDemo {
		url = UserInfoURLDemo
	}

	userInfo := UserInfo{}

	requestConfig := go_http.RequestConfig{
		URL:           url,
		ResponseModel: &userInfo,
	}
	_, _, e := service.get(&requestConfig)
	if e != nil {
		return nil, e
	}

	return &userInfo, nil
}
