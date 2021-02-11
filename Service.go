package docusign

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	errortools "github.com/leapforce-libraries/go_errortools"
	google "github.com/leapforce-libraries/go_google"
	bigquery "github.com/leapforce-libraries/go_google/bigquery"
	go_http "github.com/leapforce-libraries/go_http"
	oauth2 "github.com/leapforce-libraries/go_oauth2"
)

const (
	APIName              string = "DocuSign"
	APIURL               string = "https://demo.docusign.com/restapi/v2.1"
	APIURLDemo           string = "https://demo.docusign.net/restapi/v2.1"
	AuthURL              string = "https://account.docusign.com/oauth/auth"
	AuthURLDemo          string = "https://account-d.docusign.com/oauth/auth"
	TokenURL             string = "https://start.exactonline.nl/api/oauth2/token"
	RedirectURL          string = "http://localhost:8080/oauth/redirect"
	CustomState          string = "Leapforce!DocuSign"
	AccessTokenMethod    string = http.MethodPost
	AccessTokenGrantType string = "client_credentials"
	AccessTokenScope     string = "Elfskot.Api"
)

// Service stores Service configuration
//
type Service struct {
	integrationKey string
	isDemo         bool
	oAuth2         *oauth2.OAuth2
}

type ServiceConfig struct {
	IntegrationKey        string
	IsDemo                *bool
	MaxRetries            *uint
	SecondsBetweenRetries *uint32
}

// methods
//
func NewService(serviceConfig ServiceConfig, bigQueryService *bigquery.Service) (*Service, *errortools.Error) {
	if serviceConfig.IntegrationKey == "" {
		return nil, errortools.ErrorMessage("IntegrationKey not provided")
	}

	isDemo := false
	if serviceConfig.IsDemo != nil {
		isDemo = *serviceConfig.IsDemo
	}
	service := Service{
		integrationKey: serviceConfig.IntegrationKey,
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

func (service *Service) InitToken() *errortools.Error {
	values := url.Values{}
	values.Set("response_type", "code")
	values.Set("scope", "")
	values.Set("client_id", service.integrationKey)
	values.Set("state", CustomState)
	values.Set("redirect_uri", RedirectURL)

	fmt.Println("Go to this url to get new access token:\n")
	fmt.Printf("%s?%s\n", AuthURLDemo, values.Encode())

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
		_ = code

		b, _ := ioutil.ReadAll(r.Response.Body)

		fmt.Println(string(b))

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

func ParseDateString(date string) *time.Time {
	if len(date) >= 19 {
		d, err := time.Parse("2006-01-02T15:04:05", date[:19])
		if err == nil {
			return &d
		}
	}

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
