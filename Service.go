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
	APIName      string = "DocuSign"
	APIURL       string = "https://demo.docusign.com/restapi/v2.1"
	APIURLDemo   string = "https://demo.docusign.net/restapi/v2.1"
	AuthURL      string = "https://account.docusign.com/oauth/auth"
	AuthURLDemo  string = "https://account-d.docusign.com/oauth/auth"
	TokenURL     string = "https://account.docusign.com/oauth/token"
	TokenURLDemo string = "https://account-d.docusign.com/oauth/token"
	RedirectURL  string = "http://localhost:8080/oauth/redirect"
	CustomState  string = "Leapforce!DocuSign"
)

// Service stores Service configuration
//
type Service struct {
	userName       string
	integrationKey string
	privateKey     string
	scopes         string
	isDemo         bool
	oAuth2         *oauth2.OAuth2
}

type ServiceConfig struct {
	UserName              string
	IntegrationKey        string
	PrivateKey            string
	Scopes                string
	IsDemo                *bool
	MaxRetries            *uint
	SecondsBetweenRetries *uint32
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
		return google.GetToken(APIName, serviceConfig.IntegrationKey, bigQueryService)
	}

	saveTokenFunction := func(token *oauth2.Token) *errortools.Error {
		return google.SaveToken(APIName, serviceConfig.IntegrationKey, token, bigQueryService)
	}

	newTokenFunction := func() (*oauth2.Token, *errortools.Error) {
		return service.GetAccessToken()
	}

	oAuth2Config := oauth2.OAuth2Config{
		GetTokenFunction:      &getTokenFunction,
		SaveTokenFunction:     &saveTokenFunction,
		NewTokenFunction:      &newTokenFunction,
		MaxRetries:            serviceConfig.MaxRetries,
		SecondsBetweenRetries: serviceConfig.SecondsBetweenRetries,
	}
	service.oAuth2 = oauth2.NewOAuth(oAuth2Config)
	return &service, nil
}

func (service *Service) ValidateToken() (*oauth2.Token, *errortools.Error) {
	return service.oAuth2.ValidateToken()
}

func (service *Service) InitToken(scopes string) *errortools.Error {
	if service == nil {
		return errortools.ErrorMessage("Service variable is nil pointer")
	}

	authURL := AuthURL
	if service.isDemo {
		authURL = AuthURLDemo
	}

	values := url.Values{}
	values.Set("client_id", service.integrationKey)
	values.Set("response_type", "code")
	values.Set("redirect_uri", RedirectURL)
	values.Set("scope", scopes)
	values.Set("state", CustomState)

	url2 := fmt.Sprintf("%s?%s", authURL, values.Encode())

	fmt.Println("Go to this url to get new access token:\n")
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

// generic Get method
//
func (service *Service) get(requestConfig *go_http.RequestConfig) (*http.Request, *http.Response, *errortools.Error) {
	return service.oAuth2.Get(requestConfig)
}

// generic Post method
//
func (service *Service) post(requestConfig *go_http.RequestConfig) (*http.Request, *http.Response, *errortools.Error) {
	return service.oAuth2.Post(requestConfig)
}

// generic Put method
//
func (service *Service) put(requestConfig *go_http.RequestConfig) (*http.Request, *http.Response, *errortools.Error) {
	return service.oAuth2.Put(requestConfig)
}

// generic Patch method
//
func (service *Service) patch(requestConfig *go_http.RequestConfig) (*http.Request, *http.Response, *errortools.Error) {
	return service.oAuth2.Patch(requestConfig)
}

// generic Delete method
//
func (service *Service) delete(requestConfig *go_http.RequestConfig) (*http.Request, *http.Response, *errortools.Error) {
	return service.oAuth2.Delete(requestConfig)
}

func (service *Service) url(path string) string {
	if service.isDemo {
		return fmt.Sprintf("%s/%s", APIURLDemo, path)
	} else {
		return fmt.Sprintf("%s/%s", APIURL, path)
	}
}

func (service *Service) httpRequest(httpMethod string, requestConfig *go_http.RequestConfig, skipAccessToken bool) (*http.Request, *http.Response, *errortools.Error) {
	errorResponse := ErrorResponse{}
	(*requestConfig).ErrorModel = &errorResponse

	request, response, e := service.oAuth2.HTTPRequest(httpMethod, requestConfig, skipAccessToken)

	if e != nil {
		if errorResponse.Message != "" {
			e.SetMessage(errorResponse.Message)
		}

		b, _ := json.Marshal(errorResponse)
		e.SetExtra("error", string(b))
	}

	return request, response, e
}
