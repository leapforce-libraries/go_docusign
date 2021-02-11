package docusign

// ErrorResponse stores general DocuSign error response
//
type ErrorResponse struct {
	Error   string `json:"errorCode"`
	Message string `json:"message"`
}
