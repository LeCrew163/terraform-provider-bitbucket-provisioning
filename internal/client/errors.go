package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// BitbucketError represents an error response from the Bitbucket API
type BitbucketError struct {
	StatusCode int
	Message    string
	Context    string
	Errors     []BitbucketErrorDetail
}

// BitbucketErrorDetail represents a detailed error from the API
type BitbucketErrorDetail struct {
	Context       string `json:"context"`
	Message       string `json:"message"`
	ExceptionName string `json:"exceptionName"`
}

// Error implements the error interface
func (e *BitbucketError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("Bitbucket API error (status %d): %s", e.StatusCode, e.Message)
	}
	return fmt.Sprintf("Bitbucket API error (status %d)", e.StatusCode)
}

// ParseErrorResponse parses an HTTP error response into a BitbucketError
func ParseErrorResponse(resp *http.Response) error {
	if resp == nil {
		return fmt.Errorf("nil response")
	}

	bitbucketErr := &BitbucketError{
		StatusCode: resp.StatusCode,
	}

	// Try to read and parse the response body
	if resp.Body != nil {
		defer resp.Body.Close()
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil && len(bodyBytes) > 0 {
			// Try to parse as JSON error response
			var apiError struct {
				Errors []BitbucketErrorDetail `json:"errors"`
			}
			if err := json.Unmarshal(bodyBytes, &apiError); err == nil && len(apiError.Errors) > 0 {
				bitbucketErr.Errors = apiError.Errors
				if len(apiError.Errors) > 0 {
					bitbucketErr.Message = apiError.Errors[0].Message
					bitbucketErr.Context = apiError.Errors[0].Context
				}
			} else {
				// Use raw body as message if JSON parsing fails
				bitbucketErr.Message = string(bodyBytes)
			}
		}
	}

	// Provide default messages for common status codes
	if bitbucketErr.Message == "" {
		bitbucketErr.Message = getDefaultErrorMessage(resp.StatusCode)
	}

	return bitbucketErr
}

// getDefaultErrorMessage returns a default error message for HTTP status codes
func getDefaultErrorMessage(statusCode int) string {
	switch statusCode {
	case http.StatusUnauthorized:
		return "Authentication failed. Please check your credentials."
	case http.StatusForbidden:
		return "Permission denied. You don't have access to this resource."
	case http.StatusNotFound:
		return "Resource not found."
	case http.StatusConflict:
		return "Resource already exists or conflicts with existing resource."
	case http.StatusTooManyRequests:
		return "Rate limit exceeded. Please try again later."
	case http.StatusInternalServerError:
		return "Internal server error occurred."
	case http.StatusServiceUnavailable:
		return "Service temporarily unavailable. Please try again later."
	default:
		return "Unknown error occurred."
	}
}

// HandleError converts an error to Terraform diagnostics
func HandleError(summary string, err error) diag.Diagnostics {
	var diags diag.Diagnostics

	if err == nil {
		return diags
	}

	// Check if it's a Bitbucket API error
	if bitbucketErr, ok := err.(*BitbucketError); ok {
		detail := bitbucketErr.Message

		// Add context and additional error details if available
		if bitbucketErr.Context != "" {
			detail = fmt.Sprintf("%s\nContext: %s", detail, bitbucketErr.Context)
		}

		if len(bitbucketErr.Errors) > 1 {
			detail += "\n\nAdditional errors:"
			for i, e := range bitbucketErr.Errors[1:] {
				detail += fmt.Sprintf("\n%d. %s", i+2, e.Message)
				if e.Context != "" {
					detail += fmt.Sprintf(" (context: %s)", e.Context)
				}
			}
		}

		diags.AddError(summary, detail)
		return diags
	}

	// Generic error
	diags.AddError(summary, err.Error())
	return diags
}

// HandleWarning converts a warning to Terraform diagnostics
func HandleWarning(summary string, detail string) diag.Diagnostics {
	var diags diag.Diagnostics
	diags.AddWarning(summary, detail)
	return diags
}

// IsNotFound checks if an error indicates a resource was not found
func IsNotFound(err error) bool {
	if bitbucketErr, ok := err.(*BitbucketError); ok {
		return bitbucketErr.StatusCode == http.StatusNotFound
	}
	return false
}

// IsConflict checks if an error indicates a resource conflict
func IsConflict(err error) bool {
	if bitbucketErr, ok := err.(*BitbucketError); ok {
		return bitbucketErr.StatusCode == http.StatusConflict
	}
	return false
}

// IsUnauthorized checks if an error indicates authentication failure
func IsUnauthorized(err error) bool {
	if bitbucketErr, ok := err.(*BitbucketError); ok {
		return bitbucketErr.StatusCode == http.StatusUnauthorized
	}
	return false
}

// IsForbidden checks if an error indicates permission denial
func IsForbidden(err error) bool {
	if bitbucketErr, ok := err.(*BitbucketError); ok {
		return bitbucketErr.StatusCode == http.StatusForbidden
	}
	return false
}
