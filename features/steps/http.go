package steps

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/cucumber/godog"
	"github.com/google/go-cmp/cmp"
)

type HTTPSteps struct {
	ApplicationMux func() http.Handler

	Handler http.Handler

	ExtraHeaders http.Header
	Request      *http.Request
	Response     *httptest.ResponseRecorder
}

func (s *HTTPSteps) InitializeSuite(suite *godog.TestSuiteContext) error {
	return nil
}

func (s *HTTPSteps) InitializeScenario(scenario *godog.ScenarioContext) error {
	scenario.Before(func(ctx context.Context, sc *godog.Scenario) (context.Context, error) {
		s.ExtraHeaders = make(http.Header)
		s.Request = nil
		s.Response = nil
		return ctx, nil
	})

	scenario.Step(`^the client's remote address is "([^"]+)"$`, s.GivenClientRemoteAddr)

	scenario.Step(`^the client does a ([^ ]*) request to "([^"]+)"$`, s.WhenClientRequests)
	scenario.Step(`^the client does a ([^ ]*) request to "([^"]+)" with the following data:$`, s.WhenClientRequestsWithData)
	scenario.Step(`^the client does a ([^ ]*) request to "([^"]+)" with the following headers:$`, s.WhenClientRequestsWithHeaders)

	scenario.Step(`^the response code should be (\d+) \([^\)]+\)$`, s.ThenStatusShouldBe)
	scenario.Step(`^the response header "([^"]*)" should be "([^"]*)"$`, s.ThenHeaderShouldBe)
	scenario.Step(`^the response header "([^"]*)" should be not set$`, s.ThenHeaderShouldBeNotSet)
	scenario.Step(`^the response body should be the following "([^"]+)":$`, s.ThenResponseBodyShouldBe)
	scenario.Step(`^the response body should be empty$`, s.ThenResponseBodyShouldBeEmpty)

	return nil
}

func (s *HTTPSteps) determinHandler() http.Handler {
	if s.Handler != nil {
		return s.Handler
	}
	if s.ApplicationMux != nil {
		return s.ApplicationMux()
	}
	return nil
}

func (s *HTTPSteps) doRequest(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			if err != http.ErrAbortHandler {
				panic(err)
			}
		}
	}()
	for k, v := range s.ExtraHeaders {
		s.Request.Header[k] = v
	}
	handler := s.determinHandler()
	if handler == nil {
		return fmt.Errorf("no handler was set")
	}
	handler.ServeHTTP(s.Response, s.Request.WithContext(ctx))
	return nil
}

func (s *HTTPSteps) GivenClientRemoteAddr(ctx context.Context, addr string) error {
	s.ExtraHeaders.Set("X-Forwarded-For", addr)
	return nil
}

func (s *HTTPSteps) WhenClientRequests(ctx context.Context, method, path string) error {
	s.Request = httptest.NewRequest(method, path, nil)
	s.Response = httptest.NewRecorder()
	return s.doRequest(ctx)
}

func (s *HTTPSteps) WhenClientRequestsWithHeaders(ctx context.Context, method, path string, headers *godog.Table) error {
	s.Request = httptest.NewRequest(method, path, nil)
	for _, row := range headers.Rows {
		s.Request.Header.Set(row.Cells[0].Value, row.Cells[1].Value)
	}
	s.Response = httptest.NewRecorder()
	return s.doRequest(ctx)
}

func (s *HTTPSteps) WhenClientRequestsWithData(ctx context.Context, method, path string, data *godog.DocString) error {
	body := io.NopCloser(strings.NewReader(data.Content))
	s.Request = httptest.NewRequest(method, path, body)
	s.Response = httptest.NewRecorder()
	return s.doRequest(ctx)
}

func (s *HTTPSteps) ThenStatusShouldBe(ctx context.Context, status int) error {
	if s.Response == nil {
		return fmt.Errorf("no request was made")
	}
	if s.Response.Code != status {
		body := strings.TrimSpace(s.Response.Body.String())
		return fmt.Errorf("expected status %d, got %d (body: %v)", status, s.Response.Code, body)
	}
	return nil
}

func (s *HTTPSteps) ThenHeaderShouldBe(ctx context.Context, key, value string) error {
	if s.Response == nil {
		return fmt.Errorf("no request was made")
	}
	got := s.Response.Header().Get(key)
	if got != value {
		return fmt.Errorf("expected header %q to be %q, got %q", key, value, got)
	}
	return nil
}

func (s *HTTPSteps) ThenHeaderShouldBeNotSet(ctx context.Context, key string) error {
	if s.Response == nil {
		return fmt.Errorf("no request was made")
	}
	got, found := s.Response.Header()[key]
	if found {
		return fmt.Errorf("expected header %q to be not set, got %q", key, got)
	}
	return nil
}

func (s *HTTPSteps) ThenResponseBodyShouldBe(ctx context.Context, expectedContentType string, body *godog.DocString) error {
	if s.Response == nil {
		return fmt.Errorf("no request was made")
	}
	contentType := s.Response.Header().Get("Content-Type")
	if contentType != expectedContentType {
		body := strings.TrimSpace(s.Response.Body.String())
		return fmt.Errorf("expected content type %q, got %q (%v)", expectedContentType, contentType, body)
	}

	got := s.Response.Body.String()
	want := body.Content

	transformJSON := cmp.FilterValues(func(x, y string) bool {
		return json.Valid([]byte(x)) && json.Valid([]byte(y))
	}, cmp.Transformer("ParseJSON", func(in string) (out interface{}) {
		if err := json.Unmarshal([]byte(in), &out); err != nil {
			panic(err)
		}
		return out
	}))

	if diff := cmp.Diff(want, got, transformJSON); diff != "" {
		return fmt.Errorf("body mismatch (-want +got):\n%s", diff)
	}
	return nil
}

func (s *HTTPSteps) ThenResponseBodyShouldBeEmpty(ctx context.Context) error {
	if s.Response == nil {
		return fmt.Errorf("no request was made")
	}
	if got := s.Response.Body.String(); got != "" {
		return fmt.Errorf("expected empty body, got %q", got)
	}
	return nil
}
