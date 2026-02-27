package provider_test

import (
	"net/http"
	"os"
	"regexp"
	"time"
)

// addBasicAuth sets the Authorization header using credentials from environment
// variables. BITBUCKET_TOKEN takes precedence over username/password.
func addBasicAuth(r *http.Request) {
	if token := os.Getenv("BITBUCKET_TOKEN"); token != "" {
		r.SetBasicAuth(token, "")
		return
	}
	r.SetBasicAuth(os.Getenv("BITBUCKET_USERNAME"), os.Getenv("BITBUCKET_PASSWORD"))
}

// testBitbucketBaseURL returns the Bitbucket base URL from the environment.
func testBitbucketBaseURL() string {
	return os.Getenv("BITBUCKET_BASE_URL")
}

// newTestHTTPClient returns an HTTP client configured for acceptance test API calls.
func newTestHTTPClient() *http.Client {
	return &http.Client{Timeout: 10 * time.Second}
}

// newTestRequest builds an HTTP request with Bitbucket credentials from the environment.
func newTestRequest(method, url string) *http.Request {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		panic("testhelpers: failed to build request: " + err.Error())
	}
	addBasicAuth(req)
	return req
}

// errorRegexp is a convenience wrapper around regexp.MustCompile.
func errorRegexp(pattern string) *regexp.Regexp {
	return regexp.MustCompile(pattern)
}
