package organisations

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	server           *httptest.Server
	organisationsURL string
	isFound          bool
)

const (
	expectedCacheControlHeader string = "special header"
	canonicalUUID              string = "00000000-0000-002a-0000-00000000002a"
	alternateUUID              string = "00000000-0000-002a-0000-00000000002b"
)

type mockHTTPClient struct {
	resp       string
	statusCode int
	err        error
}

type testCase struct {
	name         string
	url          string
	clientCode   int
	clientBody   string
	clientError  error
	expectedCode int
	expectedBody string
}

func (mhc *mockHTTPClient) Do(req *http.Request) (resp *http.Response, err error) {
	cb := ioutil.NopCloser(bytes.NewReader([]byte(mhc.resp)))
	return &http.Response{Body: cb, StatusCode: mhc.statusCode}, mhc.err
}

func TestHandlers(t *testing.T) {
	var mockClient mockHTTPClient
	router := mux.NewRouter()

	invalidUUID := testCase{
		"Get organisation - Invalid UUID results in error",
		"/organisations/1234",
		200,
		getBasicOrganisationAsConcept,
		nil,
		400,
		`{"message": "uuid '1234' is either missing or invalid"}`,
	}
	conceptApiError := testCase{
		"Get organisations - Concepts API Error results in error",
		"/organisations/2d3e16e0-61cb-4322-8aff-3b01c59f4daa",
		503,
		"",
		errors.New("Downstream error"),
		500,
		`{"message": "failed to return organisation"}`,
	}
	redirectedUUID := testCase{
		"Get organisations - Given UUID was not canonical",
		"/organisations/2d3e16e0-61cb-4322-8aff-3b01c59f4daa",
		200,
		getRedirectedOrganisation,
		nil,
		301,
		``,
	}
	errorOnInvalidJson := testCase{
		"Get organisations - Error on invalid json",
		"/organisations/52aa645b-79d6-4f6f-910b-e1cff3f25a15",
		200,
		`{`,
		nil,
		500,
		`{"message": "failed to return organisation"}`,
	}
	notFound := testCase{
		"Get organisation - not found",
		"/organisations/2d3e16e0-61cb-4322-8aff-3b01c59f4daa",
		404,
		"",
		nil,
		404,
		`{"message": "organisation not found"}`,
	}
	nonOrganisationsReturnsNotFound := testCase{
		"Get organisation - Other type returns not found",
		"/organisations/f92a4ca4-84f9-11e8-8f42-da24cd01f044",
		200,
		getPersonAsConcept,
		nil,
		404,
		`{"message": "organisation not found"}`,
	}
	// TODO: Need to test...financial instrument, subsidiaries, parentOrganisation, formerNames, hortNames
	// successfulRequest := testCase{
	// 	"Get organisation - Retrieves and transforms correctly",
	// 	"/organisations/9636919c-838d-11e8-8f42-da24cd01f044",
	// 	200,
	// 	getCompleteBrandAsConcept,
	// 	nil,
	// 	200,
	// 	transformedCompleteBrand,
	// }

	testCases := []testCase{
		invalidUUID,
		conceptApiError,
		redirectedUUID,
		errorOnInvalidJson,
		notFound,
		nonOrganisationsReturnsNotFound,
		// successfulRequest,
	}

	for _, test := range testCases {
		mockClient.resp = test.clientBody
		mockClient.statusCode = test.clientCode
		mockClient.err = test.clientError
		bh := NewHandler(&mockClient, "localhost:8080/concepts")
		bh.RegisterHandlers(router)

		rr := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", test.url, nil)

		router.ServeHTTP(rr, req)
		assert.Equal(t, test.expectedCode, rr.Code, test.name+" failed: status codes do not match!")

		if rr.Code == 200 {
			assert.Equal(t, transformBody(test.expectedBody), rr.Body.String(), test.name+" failed: status body does not match!")
			continue
		}
		assert.Equal(t, test.expectedBody, rr.Body.String(), test.name+" failed: status body does not match!")
	}
}

func TestHeadersOKOnFoundForCanonicalNode(t *testing.T) {
	var mockClient mockHTTPClient
	mockClient.resp = getBasicOrganisationAsConcept
	mockClient.statusCode = 200
	mockClient.err = nil

	CacheControlHeader = expectedCacheControlHeader

	router := mux.NewRouter()
	bh := NewHandler(&mockClient, "localhost:8080/concepts")
	bh.RegisterHandlers(router)

	rec := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/organisations/d6b12f0c-bf3f-4045-a07b-1e4e49103fd6", nil)

	router.ServeHTTP(rec, req)
	fmt.Print(rec.Body)

	assert.Equal(t, expectedCacheControlHeader, rec.Header().Get("Cache-Control"))
	assert.Equal(t, "application/json; charset=UTF-8", rec.Header().Get("Content-Type"))
}

func transformBody(testBody string) string {
	stripNewLines := strings.Replace(testBody, "\n", "", -1)
	stripTabs := strings.Replace(stripNewLines, "\t", "", -1)
	return stripTabs + "\n"
}

var getBasicOrganisationAsConcept = `{
	"id": "http://api.ft.com/things/d6b12f0c-bf3f-4045-a07b-1e4e49103fd6",
	"apiUrl": "http://api.ft.com/concepts/d6b12f0c-bf3f-4045-a07b-1e4e49103fd6",
	"type": "http://www.ft.com/ontology/organisation/Organisation",
	"prefLabel": "Google Inc"
}`

var getRedirectedOrganisation = `{
	"id": "http://api.ft.com/things/d6b12f0c-bf3f-4045-a07b-1e4e49103fd6",
	"apiUrl": "http://api.ft.com/concepts/d6b12f0c-bf3f-4045-a07b-1e4e49103fd6",
	"type": "http://www.ft.com/ontology/organisation/Organisation",
	"prefLabel": "Google Inc"
}`

var getPersonAsConcept = `{
	"id": "http://api.ft.com/things/f92a4ca4-84f9-11e8-8f42-da24cd01f044",
	"apiUrl": "http://api.ft.com/concepts/f92a4ca4-84f9-11e8-8f42-da24cd01f044",
	"type": "http://www.ft.com/ontology/person/Person",
	"prefLabel": "Not a organisation"
}`
