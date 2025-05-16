package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDynamicPathHandler(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid path",
			path:           "/foo/bar/alpha123/baz/beta456/qux",
			expectedStatus: http.StatusOK,
			expectedBody:   "Path parameters received:\nFirst parameter: alpha123\nSecond parameter: beta456\n",
		},
		{
			name:           "invalid path - too few segments",
			path:           "/foo/bar/alpha123/baz/qux",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "invalid path - non-alphanumeric param1",
			path:           "/foo/bar/alpha-123/baz/beta456/qux",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "invalid path - non-alphanumeric param2",
			path:           "/foo/bar/alpha123/baz/beta_456/qux",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "path prefix matches but full path does not",
			path:           "/foo/bar/not/enough/segments",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "invalid path - `baz` segment missing",
			path:           "/foo/bar/paramOne/paramTwo/qux",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "invalid path - `qux` segment missing",
			path:           "/foo/bar/paramOne/baz/paramTwo",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "invalid path - format correct with different segments",
			path:           "/foo/bar/paramOne/baaz/paramTwo/quux",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			// For dynamicPathHandler, it's registered with a prefix,
			// so we simulate a mux that would call it.
			// A direct call is simpler here as the handler itself contains the full logic.
			handler := http.HandlerFunc(dynamicPathHandler)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(tt.expectedBody) {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestCustomRouter(t *testing.T) {
	router := &customRouter{}
	router.HandleFunc("/foo/bar/%s/baz/%s/qux", dynamicPathHandlerFunc)

	tests := []struct {
		name           string
		path           string
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "valid path",
			path:           "/foo/bar/paramOne/baz/paramTwo/qux",
			expectedStatus: http.StatusOK,
			expectedBody:   "Path parameters received:\nFirst parameter: paramOne\nSecond parameter: paramTwo\n",
		},
		{
			name:           "valid path with alphanumeric params",
			path:           "/foo/bar/param123/baz/param456/qux",
			expectedStatus: http.StatusOK,
			expectedBody:   "Path parameters received:\nFirst parameter: param123\nSecond parameter: param456\n",
		},
		{
			name:           "invalid path - `baz` segment missing",
			path:           "/foo/bar/paramOne/paramTwo/qux",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "invalid path - `qux` segment missing",
			path:           "/foo/bar/paramOne/baz/paramTwo",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "invalid path - format correct with different segments",
			path:           "/foo/bar/paramOne/baaz/paramTwo/quux",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "invalid path - too few segments - missing last segment",
			path:           "/foo/bar/paramOne/baz/qux",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "invalid path - pattern mismatch",
			path:           "/foo/bar/paramOne/wrong/paramTwo/qux",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "invalid path - non-alphanumeric param1",
			path:           "/foo/bar/param-One/baz/paramTwo/qux",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "invalid path - non-alphanumeric param2",
			path:           "/foo/bar/paramOne/baz/param_Two/qux",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(tt.expectedBody) {
				t.Errorf("handler returned unexpected body: got %v want %v",
					rr.Body.String(), tt.expectedBody)
			}
		})
	}
}
