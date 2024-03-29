package docusign

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	errortools "github.com/leapforce-libraries/go_errortools"
	google "github.com/leapforce-libraries/go_google"
	bigquery "github.com/leapforce-libraries/go_google/bigquery"
	go_http "github.com/leapforce-libraries/go_http"
	oauth2 "github.com/leapforce-libraries/go_oauth2"
)

const (
	apiName         string = "DocuSign"
	userInfoURL     string = "https://account.docusign.com/oauth/userinfo"
	userInfoURLDemo string = "https://account-d.docusign.com/oauth/userinfo"
	authURL         string = "https://account.docusign.com/oauth/auth"
	authURLDemo     string = "https://account-d.docusign.com/oauth/auth"
	tokenURL        string = "https://account.docusign.com/oauth/token"
	tokenURLDemo    string = "https://account-d.docusign.com/oauth/token"
	redirectURL     string = "http://localhost:8080/oauth/redirect"
	customState     string = "Leapforce!DocuSign"
)

// Service stores Service configuration
//
type Service struct {
	userName       string
	integrationKey string
	privateKey     string
	scopes         string
	isDemo         bool
	userInfo       *UserInfo
	baseURIs       map[string]string
	oAuth2Service  *oauth2.Service
}
type ServiceConfig struct {
	UserName       string
	IntegrationKey string
	PrivateKey     string
	Scopes         string
	IsDemo         *bool
}

// methods
//
func NewService(serviceConfig ServiceConfig, bigQueryService *bigquery.Service) (*Service, *errortools.Error) {
	if serviceConfig.UserName == "" {
		return nil, errortools.ErrorMessage("UserName not provided")
	}
	if serviceConfig.IntegrationKey == "" {
		return nil, errortools.ErrorMessage("IntegrationKey not provided")
	}
	if serviceConfig.PrivateKey == "" {
		return nil, errortools.ErrorMessage("PrivateKey not provided")
	}

	isDemo := false
	if serviceConfig.IsDemo != nil {
		isDemo = *serviceConfig.IsDemo
	}
	service := Service{
		userName:       serviceConfig.UserName,
		integrationKey: serviceConfig.IntegrationKey,
		privateKey:     serviceConfig.PrivateKey,
		scopes:         serviceConfig.Scopes,
		isDemo:         isDemo,
	}

	getTokenFunction := func() (*oauth2.Token, *errortools.Error) {
		return google.GetToken(apiName, serviceConfig.IntegrationKey, bigQueryService)
	}

	saveTokenFunction := func(token *oauth2.Token) *errortools.Error {
		return google.SaveToken(apiName, serviceConfig.IntegrationKey, token, bigQueryService)
	}

	newTokenFunction := func() (*oauth2.Token, *errortools.Error) {
		return service.GetAccessToken()
	}

	oAuth2ServiceConfig := oauth2.ServiceConfig{
		GetTokenFunction:  &getTokenFunction,
		SaveTokenFunction: &saveTokenFunction,
		NewTokenFunction:  &newTokenFunction,
	}
	oAuth2Service, e := oauth2.NewService(&oAuth2ServiceConfig)
	if e != nil {
		return nil, e
	}

	service.oAuth2Service = oAuth2Service
	return &service, nil
}

func (service *Service) ValidateToken() (*oauth2.Token, *errortools.Error) {
	return service.oAuth2Service.ValidateToken()
}

func (service *Service) InitToken(scopes string) *errortools.Error {
	if service == nil {
		return errortools.ErrorMessage("Service variable is nil pointer")
	}

	authURL := authURL
	if service.isDemo {
		authURL = authURLDemo
	}

	values := url.Values{}
	values.Set("client_id", service.integrationKey)
	values.Set("response_type", "code")
	values.Set("redirect_uri", redirectURL)
	values.Set("scope", scopes)
	values.Set("state", customState)

	url2 := fmt.Sprintf("%s?%s", authURL, values.Encode())

	fmt.Println("Go to this url to get new access token:")
	fmt.Println(url2 + "\n")

	// Create a new redirect route
	http.HandleFunc("/oauth/redirect", func(w http.ResponseWriter, r *http.Request) {
		//
		// get authorization code
		//
		err := r.ParseForm()
		if err != nil {
			fmt.Fprintf(os.Stdout, "could not parse query: %v", err)
			w.WriteHeader(http.StatusBadRequest)
		}
		code := r.FormValue("code")

		fmt.Println(code)

		/*ee := oa.getTokenFromCode(code)
		if ee != nil {
			fmt.Println(ee.Message)
		}*/

		w.WriteHeader(http.StatusFound)

		return
	})

	http.ListenAndServe(":8080", nil)

	return nil
}

func (service *Service) url(accountID string, path string) (string, *errortools.Error) {
	if service.baseURIs == nil {
		// get userinfo
		userInfo, e := service.GetUserInfo()
		if e != nil {
			return "", e
		}
		service.userInfo = userInfo
		service.baseURIs = make(map[string]string)
	}

	baseURI, ok := service.baseURIs[accountID]
	if ok {
		goto found
	}

	for _, account := range service.userInfo.Accounts {
		if account.AccountID == accountID {
			baseURI = account.BaseURI
			service.baseURIs[accountID] = baseURI
			goto found
		}
	}

	return "", errortools.ErrorMessage(fmt.Sprintf("UserInfo does not contain info for account '%s'", accountID))
found:

	return fmt.Sprintf("%s/restapi/v2.1/%s", baseURI, path), nil
}

func (service *Service) httpRequest(requestConfig *go_http.RequestConfig) (*http.Request, *http.Response, *errortools.Error) {
	errorResponse := ErrorResponse{}
	(*requestConfig).ErrorModel = &errorResponse

	request, response, e := service.oAuth2Service.HTTPRequest(requestConfig)

	if e != nil {
		if errorResponse.Message != "" {
			e.SetMessage(errorResponse.Message)
		}

		b, _ := json.Marshal(errorResponse)
		e.SetExtra("error", string(b))
	}

	return request, response, e
}

func (service *Service) httpRequestWithoutAccessToken(requestConfig *go_http.RequestConfig) (*http.Request, *http.Response, *errortools.Error) {
	errorResponse := ErrorResponse{}
	(*requestConfig).ErrorModel = &errorResponse

	request, response, e := service.oAuth2Service.HTTPRequestWithoutAccessToken(requestConfig)

	if e != nil {
		if errorResponse.Message != "" {
			e.SetMessage(errorResponse.Message)
		}

		b, _ := json.Marshal(errorResponse)
		e.SetExtra("error", string(b))
	}

	return request, response, e
}
