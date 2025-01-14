package middleware

import (
	"context"
	"fmt"
	"github.com/CMSgov/dpc/api/constants"
	"github.com/kinbiko/jsonassert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ContextTestSuite struct {
	suite.Suite
}

func TestContextTestSuite(t *testing.T) {
	suite.Run(t, new(ContextTestSuite))
}

func (suite *ContextTestSuite) TestExportTypesParam() {
	var types string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		types, _ = r.Context().Value(constants.ContextKeyResourceTypes).(string)
	})

	req := httptest.NewRequest(http.MethodGet, "http://www.example.com/Group/some-id/$export?_type=Coverage,ExplanationOfBenefit&_since=2012-01-02T12:12:12-05:00", nil)
	res := httptest.NewRecorder()

	e := ExportTypesParamCtx(nextHandler)
	e.ServeHTTP(res, req)
	assert.Equal(suite.T(), "Coverage,ExplanationOfBenefit", types)
}

func (suite *ContextTestSuite) TestExportTypesParamDefault() {
	var types string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		types, _ = r.Context().Value(constants.ContextKeyResourceTypes).(string)
	})

	req := httptest.NewRequest(http.MethodGet, "http://www.example.com/Group/some-id/$export?_since=2012-01-02T12:12:12-05:00", nil)
	res := httptest.NewRecorder()

	e := ExportTypesParamCtx(nextHandler)
	e.ServeHTTP(res, req)
	assert.Equal(suite.T(), constants.AllResources, types)
}

func (suite *ContextTestSuite) TestExportTypesParamInvalid() {
	var types string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		types, _ = r.Context().Value(constants.ContextKeyResourceTypes).(string)
	})

	req := httptest.NewRequest(http.MethodGet, "http://www.example.com/Group/some-id/$export?_type=foobar", nil)
	res := httptest.NewRecorder()

	e := ExportTypesParamCtx(nextHandler)
	e.ServeHTTP(res, req)
	assert.Equal(suite.T(), "", types)
}

func (suite *ContextTestSuite) TestExportSinceParamMinusTZ() {
	var since string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		since, _ = r.Context().Value(constants.ContextKeySince).(string)
	})

	req := httptest.NewRequest(http.MethodGet, "http://www.example.com/Group/some-id/$export?_type=Coverage,ExplanationOfBenefit&_since=2012-01-02T12:12:12-05:00", nil)
	res := httptest.NewRecorder()

	e := ExportSinceParamCtx(nextHandler)
	e.ServeHTTP(res, req)
	assert.Equal(suite.T(), "2012-01-02T12:12:12-05:00", since)
}

func (suite *ContextTestSuite) TestExportSinceParamPlusTZ() {
	var since string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		since, _ = r.Context().Value(constants.ContextKeySince).(string)
	})

	req := httptest.NewRequest(http.MethodGet, "http://www.example.com/Group/some-id/$export?_type=Coverage,ExplanationOfBenefit&_since=2012-01-02T12:12:12+05:00", nil)
	res := httptest.NewRecorder()

	e := ExportSinceParamCtx(nextHandler)
	e.ServeHTTP(res, req)
	assert.Equal(suite.T(), "", since)
}

func (suite *ContextTestSuite) TestExportSinceParamPlusTZEncoded() {
	var since string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		since, _ = r.Context().Value(constants.ContextKeySince).(string)
	})

	req := httptest.NewRequest(http.MethodGet, "http://www.example.com/Group/some-id/$export?_type=Coverage,ExplanationOfBenefit&_since=2012-01-02T12:12:12%2b05:00", nil)
	res := httptest.NewRecorder()

	e := ExportSinceParamCtx(nextHandler)
	e.ServeHTTP(res, req)
	assert.Equal(suite.T(), "2012-01-02T12:12:12+05:00", since)
}

func (suite *ContextTestSuite) TestExportSinceParamDefault() {
	var since string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		since, _ = r.Context().Value(constants.ContextKeySince).(string)
	})

	req := httptest.NewRequest(http.MethodGet, "http://www.example.com/Group/some-id/$export?_type=Coverage,ExplanationOfBenefit", nil)
	res := httptest.NewRecorder()

	e := ExportSinceParamCtx(nextHandler)
	e.ServeHTTP(res, req)
	assert.Equal(suite.T(), "", since)
}

func (suite *ContextTestSuite) TestExportSinceParamFuture() {
	future := time.Now().Add(time.Hour * 3).Format(constants.SinceLayout)
	url := fmt.Sprintf("http://www.example.com/Group/some-id/$export?_type=Coverage,ExplanationOfBenefit&_since=%s", future)
	var since string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		since, _ = r.Context().Value(constants.ContextKeySince).(string)
	})

	req := httptest.NewRequest(http.MethodGet, url, nil)
	res := httptest.NewRecorder()

	e := ExportSinceParamCtx(nextHandler)
	e.ServeHTTP(res, req)
	assert.Equal(suite.T(), "", since)
}

func (suite *ContextTestSuite) TestExportSinceParamInvalid() {
	var since string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		since, _ = r.Context().Value(constants.ContextKeySince).(string)
	})

	req := httptest.NewRequest(http.MethodGet, "http://www.example.com/Group/some-id/$export?_type=Coverage,ExplanationOfBenefit&_since=foobar", nil)
	res := httptest.NewRecorder()

	e := ExportSinceParamCtx(nextHandler)
	e.ServeHTTP(res, req)
	assert.Equal(suite.T(), "", since)
}

func (suite *ContextTestSuite) TestProvenanceHeader() {
	ja := jsonassert.New(suite.T())
	var header string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header, _ = r.Context().Value(constants.ContextKeyProvenanceHeader).(string)
	})

	t := time.Now().Format(constants.SinceLayout)
	req := httptest.NewRequest(http.MethodGet, "http://www.example.com/", nil)
	req.Header.Set(constants.ProvenanceHeader, fmt.Sprintf("{\"resourceType\":\"Provenance\",\"recorded\":\"%s\",\"reason\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/ActReason\",\"code\":\"TREAT\"}]}],\"agent\":[{\"role\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/RoleClass\",\"code\":\"AGNT\"}]}],\"who\":{\"reference\":\"Organization/c5a40867-011a-43f9-996e-aa92207fbbe2\"}}]}", t))
	res := httptest.NewRecorder()

	e := ProvenanceHeaderValidator(false)(nextHandler)
	e.ServeHTTP(res, req)
	body, _ := ioutil.ReadAll(res.Body)
	ja.Assertf(string(body), `{
      "issue": [
        {
          "severity": "error",
          "code": "Exception",
          "details": {
            "text": "Internal Server Error"
          },
          "diagnostics": "<<PRESENCE>>"
        }
      ],
      "resourceType": "OperationOutcome"
    }`)

	t = time.Now().Add(-48 * time.Hour).Format(constants.SinceLayout)
	req = httptest.NewRequest(http.MethodGet, "http://www.example.com/", nil)
	req.Header.Set(constants.ProvenanceHeader, fmt.Sprintf("{\"resourceType\":\"Provenance\",\"recorded\":\"%s\",\"reason\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/ActReason\",\"code\":\"TREAT\"}]}],\"agent\":[{\"role\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/RoleClass\",\"code\":\"AGNT\"}]}],\"who\":{\"reference\":\"Organization/c5a40867-011a-43f9-996e-aa92207fbbe2\"}}]}", t))
	res = httptest.NewRecorder()

	e = ProvenanceHeaderValidator(false)(nextHandler)
	e.ServeHTTP(res, req)
	body, _ = ioutil.ReadAll(res.Body)
	ja.Assertf(string(body), `{
      "issue": [
        {
          "severity": "warning",
          "code": "Business Rule Violation",
          "details": {
            "text": "Recorded timestamp invalid because it's outside the 24 hr window"
          },
          "diagnostics": "<<PRESENCE>>"
        }
      ],
      "resourceType": "OperationOutcome"
    }`)

	t = time.Now().Format(constants.SinceLayout)
	req = httptest.NewRequest(http.MethodGet, "http://www.example.com/", nil)
	req.Header.Set(constants.ProvenanceHeader, fmt.Sprintf("{\"resourceType\":\"Provenance\",\"recorded\":\"%s\",\"reason\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/ActReason\",\"code\":\"MEET\"}]}],\"agent\":[{\"role\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/RoleClass\",\"code\":\"AGNT\"}]}],\"who\":{\"reference\":\"Organization/c5a40867-011a-43f9-996e-aa92207fbbe2\"}}]}", t))
	res = httptest.NewRecorder()

	e = ProvenanceHeaderValidator(false)(nextHandler)
	e.ServeHTTP(res, req)
	body, _ = ioutil.ReadAll(res.Body)
	ja.Assertf(string(body), `{
      "issue": [
        {
          "severity": "warning",
          "code": "Business Rule Violation",
          "details": {
            "text": "Invalid reason"
          },
          "diagnostics": "<<PRESENCE>>"
        }
      ],
      "resourceType": "OperationOutcome"
    }`)

	t = time.Now().Format(constants.SinceLayout)
	req = httptest.NewRequest(http.MethodGet, "http://www.example.com/", nil)
	req.Header.Set(constants.ProvenanceHeader, fmt.Sprintf("{\"resourceType\":\"Provenance\",\"recorded\":\"%s\",\"reason\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/ActReason\",\"code\":\"TREAT\"}]}],\"agent\":[{\"role\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/RoleClass\",\"code\":\"AGENNT\"}]}],\"who\":{\"reference\":\"Organization/c5a40867-011a-43f9-996e-aa92207fbbe2\"}}]}", t))
	res = httptest.NewRecorder()

	e = ProvenanceHeaderValidator(false)(nextHandler)
	e.ServeHTTP(res, req)
	body, _ = ioutil.ReadAll(res.Body)
	ja.Assertf(string(body), `{
      "issue": [
        {
          "severity": "warning",
          "code": "Business Rule Violation",
          "details": {
            "text": "Invalid role"
          },
          "diagnostics": "<<PRESENCE>>"
        }
      ],
      "resourceType": "OperationOutcome"
    }`)

	t = time.Now().Format(constants.SinceLayout)
	req = httptest.NewRequest(http.MethodGet, "http://www.example.com/", nil)
	req.Header.Set(constants.ProvenanceHeader, fmt.Sprintf("{\"resourceType\":\"Provenance\",\"recorded\":\"%s\",\"reason\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/ActReason\",\"code\":\"TREAT\"}]}],\"agent\":[{\"role\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/RoleClass\",\"code\":\"AGNT\"}]}],\"who\":{\"reference\":\"Organizations/c5a40867-011a-43f9-996e-aa92207fbbe2\"}}]}", t))
	res = httptest.NewRecorder()

	e = ProvenanceHeaderValidator(false)(nextHandler)
	e.ServeHTTP(res, req)
	body, _ = ioutil.ReadAll(res.Body)
	ja.Assertf(string(body), `{
      "issue": [
        {
          "severity": "warning",
          "code": "Business Rule Violation",
          "details": {
            "text": "Invalid who reference"
          },
          "diagnostics": "<<PRESENCE>>"
        }
      ],
      "resourceType": "OperationOutcome"
    }`)

	t = time.Now().Format(constants.SinceLayout)
	req = httptest.NewRequest(http.MethodGet, "http://www.example.com/", nil)
	req.Header.Set(constants.ProvenanceHeader, fmt.Sprintf("{\"resourceType\":\"Provenance\",\"recorded\":\"%s\",\"reason\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/ActReason\",\"code\":\"TREAT\"}]}],\"agent\":[{\"role\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/RoleClass\",\"code\":\"AGNT\"}]}],\"who\":{\"reference\":\"Organization/c5a40867-011a-43f9-996e-aa92207fbbe2\"}}]}", t))
	res = httptest.NewRecorder()
	ctx := req.Context()
	ctx = context.WithValue(ctx, constants.ContextKeyOrganization, "blah blah blah")
	req = req.WithContext(ctx)

	e = ProvenanceHeaderValidator(false)(nextHandler)
	e.ServeHTTP(res, req)
	body, _ = ioutil.ReadAll(res.Body)
	ja.Assertf(string(body), `{
      "issue": [
        {
          "severity": "warning",
          "code": "Business Rule Violation",
          "details": {
            "text": "Org in Provenance not valid"
          },
          "diagnostics": "<<PRESENCE>>"
        }
      ],
      "resourceType": "OperationOutcome"
    }`)

	t = time.Now().Format(constants.SinceLayout)
	req = httptest.NewRequest(http.MethodGet, "http://www.example.com/", nil)
	req.Header.Set(constants.ProvenanceHeader, fmt.Sprintf("{\"resourceType\":\"Provenance\",\"recorded\":\"%s\",\"reason\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/ActReason\",\"code\":\"TREAT\"}]}],\"agent\":[{\"role\":[{\"coding\":[{\"system\":\"http://hl7.org/fhir/v3/RoleClass\",\"code\":\"AGNT\"}]}],\"who\":{\"reference\":\"Organization/c5a40867-011a-43f9-996e-aa92207fbbe2\"}}]}", t))
	res = httptest.NewRecorder()
	ctx = req.Context()
	ctx = context.WithValue(ctx, constants.ContextKeyOrganization, "c5a40867-011a-43f9-996e-aa92207fbbe2")
	req = req.WithContext(ctx)

	e = ProvenanceHeaderValidator(false)(nextHandler)
	e.ServeHTTP(res, req)
	body, _ = ioutil.ReadAll(res.Body)
	assert.True(suite.T(), len(body) == 0)
	assert.NotZero(suite.T(), header)

}

func (suite *ContextTestSuite) TestMBIContext() {
	var mbi string
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mbi, _ = r.Context().Value(constants.ContextKeyMBI).(string)
	})

	req := httptest.NewRequest(http.MethodGet, "http://www.example.com/doesnotmatter", nil)
	res := httptest.NewRecorder()

	e := MBICtx(nextHandler)
	e.ServeHTTP(res, req)

	assert.Equal(suite.T(), http.StatusBadRequest, res.Code)

	req = httptest.NewRequest(http.MethodGet, "http://www.example.com/doesnotmatter", nil)
	req.Header.Add(constants.FHIRIdentifierSystemHeader, "my-system")
	res = httptest.NewRecorder()

	e = MBICtx(nextHandler)
	e.ServeHTTP(res, req)

	assert.Equal(suite.T(), http.StatusBadRequest, res.Code)

	req = httptest.NewRequest(http.MethodGet, "http://www.example.com/doesnotmatter", nil)
	req.Header.Add(constants.FHIRIdentifierSystemHeader, "http://hl7.org/fhir/sid/us-mbi")
	res = httptest.NewRecorder()

	e = MBICtx(nextHandler)
	e.ServeHTTP(res, req)

	assert.Equal(suite.T(), http.StatusBadRequest, res.Code)

	req = httptest.NewRequest(http.MethodGet, "http://www.example.com/doesnotmatter", nil)
	req.Header.Add(constants.FHIRIdentifierSystemHeader, "http://hl7.org/fhir/sid/us-mbi")
	req.Header.Add(constants.FHIRIdentifierValueHeader, "mbi")
	res = httptest.NewRecorder()

	e = MBICtx(nextHandler)
	e.ServeHTTP(res, req)

	assert.Equal(suite.T(), "mbi", mbi)
}
