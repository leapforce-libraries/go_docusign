package docusign

import (
	"fmt"
	"net/http"

	errortools "github.com/leapforce-libraries/go_errortools"
	go_http "github.com/leapforce-libraries/go_http"
)

type Templates struct {
	Templates []Template `json:"envelopeTemplates"`
}

type Template struct {
	TemplateID string `json:"templateId"`
}

func (service *Service) GetTemplates(accountID string) (*Templates, *errortools.Error) {
	path := fmt.Sprintf("accounts/%s/templates", accountID)

	url, e := service.url(accountID, path)
	if e != nil {
		return nil, e
	}

	templates := Templates{}

	requestConfig := go_http.RequestConfig{
		Method:        http.MethodGet,
		URL:           url,
		ResponseModel: &templates,
	}
	_, _, e = service.httpRequest(&requestConfig)
	if e != nil {
		return nil, e
	}

	return &templates, nil
}
