package main_test

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/cucumber/godog"
	"github.com/gin-gonic/gin"
	"github.com/itsubaki/quasar/pkg/handler"
	"github.com/jfilipczyk/gomatch"
)

var (
	projectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	api       = &apiFeature{}
)

type apiFeature struct {
	header http.Header
	body   io.Reader
	resp   *httptest.ResponseRecorder

	server *gin.Engine
	keep   map[string]interface{}
}

func (a *apiFeature) start() {
	a.server = handler.New()
	a.keep = make(map[string]interface{})
}

func (a *apiFeature) reset(sc *godog.Scenario) {
	a.header = make(http.Header)
	a.body = nil
	a.resp = httptest.NewRecorder()
}

func (a *apiFeature) replace(str string) string {
	for k, v := range a.keep {
		switch val := v.(type) {
		case string:
			str = strings.Replace(str, k, val, -1)
		}
	}

	return str
}

func (a *apiFeature) Request(method, endpoint string) error {
	r := a.replace(endpoint)
	req := httptest.NewRequest(method, r, a.body)
	req.Header = a.header
	req.Header.Add(NewXCloudTraceContext())

	a.server.ServeHTTP(a.resp, req)
	return nil
}

func (a *apiFeature) ResponseCodeShouldBe(code int) error {
	if code == a.resp.Code {
		return nil
	}

	return fmt.Errorf("got=%v, want=%v", a.resp.Code, code)
}

func (a *apiFeature) ResponseShouldMatchJSON(body *godog.DocString) error {
	want := a.replace(body.Content)
	got := a.resp.Body.String()

	ok, err := gomatch.NewDefaultJSONMatcher().Match(want, got)
	if err != nil {
		return fmt.Errorf("got=%v, want=%v, match: %v", got, want, err)
	}

	if !ok {
		return fmt.Errorf("got=%v, want=%v", got, want)
	}

	return nil
}

func (a *apiFeature) SetHeader(k, v string) error {
	a.header.Add(k, v)
	return nil
}

func (a *apiFeature) SetRequestBody(body *godog.DocString) error {
	r := a.replace(body.Content)
	a.body = bytes.NewBuffer([]byte(r))
	return nil
}

func InitializeTestSuite(ctx *godog.TestSuiteContext) {
	ctx.BeforeSuite(func() {
		gin.SetMode(gin.ReleaseMode)
		api.start()
	})

	ctx.AfterSuite(func() {})
}

func InitializeScenario(ctx *godog.ScenarioContext) {
	ctx.BeforeScenario(api.reset)

	ctx.Step(`^I set "([^"]*)" header with "([^"]*)"$`, api.SetHeader)
	ctx.Step(`^I set request body:$`, api.SetRequestBody)
	ctx.Step(`^I send "([^"]*)" request to "([^"]*)"$`, api.Request)
	ctx.Step(`^the response code should be (\d+)$`, api.ResponseCodeShouldBe)
	ctx.Step(`^the response should match json:$`, api.ResponseShouldMatchJSON)
}

func NewXCloudTraceContext() (string, string) {
	return "X-Cloud-Trace-Context", fmt.Sprintf("%016x%016x/%d;", rand.Int63(), rand.Int63(), rand.Int63())
}
