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
			expectedBody:   "Path parameters received:\nParameter 1: alpha123\nParameter 2: beta456\n",
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
			pathPattern := "/foo/bar/%s/baz/%s/qux"
			handlerFunc := newDynamicPathHandler(pathPattern)
			handler := http.HandlerFunc(handlerFunc)
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

func TestNewPathRegexHandler(t *testing.T) {
	routeTemplateStr := "/foo/bar/%s/baz/%s/qux"

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
			expectedBody:   "Parameter 1: alpha123\nParameter 2: beta456\n",
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

	handler := newPathRegexHandler(routeTemplateStr)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
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

	t.Run("no parameters in route", func(t *testing.T) {
		noParamRoute := "/static/path"
		noParamHandler := newPathRegexHandler(noParamRoute)
		expectedNoParamBody := "No parameters captured.\n"

		req, err := http.NewRequest("GET", noParamRoute, nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		noParamHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expectedNoParamBody) {
			t.Errorf("handler returned unexpected body: got %q want %q", rr.Body.String(), expectedNoParamBody)
		}

		// Test mismatch for no param route
		reqMismatch, errMismatch := http.NewRequest("GET", "/static/other", nil)
		if errMismatch != nil {
			t.Fatal(errMismatch)
		}

		rrMismatch := httptest.NewRecorder()
		noParamHandler.ServeHTTP(rrMismatch, reqMismatch)
		if status := rrMismatch.Code; status != http.StatusNotFound {
			t.Errorf("handler returned wrong status code for mismatch: got %v want %v", status, http.StatusNotFound)
		}
	})

	t.Run("one parameter in route", func(t *testing.T) {
		oneParamRoute := "/user/%s/profile"
		oneParamHandler := newPathRegexHandler(oneParamRoute)
		expectedOneParamBody := "Parameter 1: testuser\n"

		req, err := http.NewRequest("GET", "/user/testuser/profile", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		oneParamHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expectedOneParamBody) {
			t.Errorf("handler returned unexpected body: got %q want %q", rr.Body.String(), expectedOneParamBody)
		}
	})

	t.Run("two parameters in route", func(t *testing.T) {
		multiParamRoute := "/foo/%s/bar/%s/baz"
		multiParamHandler := newPathRegexHandler(multiParamRoute)
		expectedMultiParamBody := "Parameter 1: param1\nParameter 2: param2\n"

		req, err := http.NewRequest("GET", "/foo/param1/bar/param2/baz", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		multiParamHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expectedMultiParamBody) {
			t.Errorf("handler returned unexpected body: got %q want %q", rr.Body.String(), expectedMultiParamBody)
		}
	})

	t.Run("three parameters in route", func(t *testing.T) {
		threeParamRoute := "/api/v3/%s/%s/%s"
		threeParamHandler := newPathRegexHandler(threeParamRoute)
		expectedThreeParamBody := "Parameter 1: param1\nParameter 2: param2\nParameter 3: param3\n"

		req, err := http.NewRequest("GET", "/api/v3/param1/param2/param3", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		threeParamHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expectedThreeParamBody) {
			t.Errorf("handler returned unexpected body: got %q want %q", rr.Body.String(), expectedThreeParamBody)
		}
	})

	t.Run("three parameters in route with other segments", func(t *testing.T) {
		threeParamRoute := "/api/v3/%s/other/%s/more/%s"
		threeParamHandler := newPathRegexHandler(threeParamRoute)
		expectedThreeParamBody := "Parameter 1: param1\nParameter 2: param2\nParameter 3: param3\n"

		req, err := http.NewRequest("GET", "/api/v3/param1/other/param2/more/param3", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		threeParamHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
		if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expectedThreeParamBody) {
			t.Errorf("handler returned unexpected body: got %q want %q", rr.Body.String(), expectedThreeParamBody)
		}
	})

	t.Run("invalid path for two parameters", func(t *testing.T) {
		multiParamRoute := "/foo/%s/bar/%s/baz"
		multiParamHandler := newPathRegexHandler(multiParamRoute)

		req, err := http.NewRequest("GET", "/foo/param1/bar/param2/invalid", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		multiParamHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusNotFound {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
		}
	})

	t.Run("invalid method", func(t *testing.T) {
		multiParamRoute := "/foo/%s/bar/%s/baz"
		multiParamHandler := newPathRegexHandler(multiParamRoute)

		req, err := http.NewRequest("POST", "/foo/param1/bar/param2/baz", nil)
		if err != nil {
			t.Fatal(err)
		}
		rr := httptest.NewRecorder()
		multiParamHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusMethodNotAllowed {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
		}

		expectedBody := "Method not allowed\n"
		if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(expectedBody) {
			t.Errorf("handler returned unexpected body: got %q want %q", rr.Body.String(), expectedBody)
		}
	})
}

func TestCustomRouter(t *testing.T) {
	router := &customRouter{}
	routeTemplateStr := "/foo/bar/%s/baz/%s/qux"
	router.HandleFunc(routeTemplateStr, newDynamicPathHandler(routeTemplateStr))

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
			expectedBody:   "Path parameters received:\nParameter 1: paramOne\nParameter 2: paramTwo\n",
		},
		{
			name:           "valid path with alphanumeric params",
			path:           "/foo/bar/param123/baz/param456/qux",
			expectedStatus: http.StatusOK,
			expectedBody:   "Path parameters received:\nParameter 1: param123\nParameter 2: param456\n",
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

func TestMakeRegexPatternStr(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{
			name:     "single %s",
			pattern:  "/api/users/%s/profile",
			expected: "^/api/users/([a-zA-Z0-9]+)/profile$",
		},
		{
			name:     "multiple %s",
			pattern:  "/foo/%s/bar/%s/baz",
			expected: "^/foo/([a-zA-Z0-9]+)/bar/([a-zA-Z0-9]+)/baz$",
		},
		{
			name:     "no %s",
			pattern:  "/static/path",
			expected: "^/static/path$",
		},
		{
			name:     "%s at the beginning",
			pattern:  "%s/path/end",
			expected: "^([a-zA-Z0-9]+)/path/end$",
		},
		{
			name:     "%s at the end",
			pattern:  "/path/start/%s",
			expected: "^/path/start/([a-zA-Z0-9]+)$",
		},
		{
			name:     "empty pattern",
			pattern:  "",
			expected: "^$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeRegexPatternStr(tt.pattern)
			if result != tt.expected {
				t.Errorf("makeRegexPatternStr(%q) = %q; want %q", tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestGetPathPrefix(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		expected string
	}{
		{
			name:     "multiple %s",
			pattern:  "/foo/bar/%s/baz/%s/qux",
			expected: "/foo/bar/",
		},
		{
			name:     "single %s",
			pattern:  "/api/users/%s/profile",
			expected: "/api/users/",
		},
		{
			name:     "multiple %s",
			pattern:  "/foo/%s/bar/%s/baz",
			expected: "/foo/",
		},
		{
			name:     "no %s",
			pattern:  "/static/path",
			expected: "/static/path",
		},
		{
			name:     "%s at the beginning",
			pattern:  "%s/path/end",
			expected: "",
		},
		{
			name:     "%s at the end with preceding slash",
			pattern:  "/path/start/%s",
			expected: "/path/start/",
		},
		{
			name:     "empty pattern",
			pattern:  "",
			expected: "",
		},
		{
			name:     "pattern is just %s",
			pattern:  "%s",
			expected: "",
		},
		{
			name:     "pattern is just /%s",
			pattern:  "/%s",
			expected: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPathPrefix(tt.pattern)
			if result != tt.expected {
				t.Errorf("getPathPrefix(%q) = %q; want %q", tt.pattern, result, tt.expected)
			}
		})
	}
}

func TestCustomRouterOverlappingRoutes(t *testing.T) {
	router := &customRouter{}
	// Define routes as in main.go for the custom router
	routeTemplates := []string{
		"/api/v3/%s/%s",
		"/api/v3/%s/%s/version", // overlapping route
	}
	router.addTemplateRoutes(routeTemplates)

	tests := []struct {
		name           string
		method         string // Added to specify HTTP method
		path           string
		expectedStatus int
		expectedBody   string
	}{
		// New tests for overlapping routes /api/v3/...
		{
			name:           "overlapping route - specific versioned path",
			path:           "/api/v3/id1/id2/version",
			expectedStatus: http.StatusOK,
			expectedBody:   "Path parameters received:\nParameter 1: id1\nParameter 2: id2\n",
		},
		{
			name:           "overlapping route - general path",
			path:           "/api/v3/idAlpha/idBeta",
			expectedStatus: http.StatusOK,
			expectedBody:   "Path parameters received:\nParameter 1: idAlpha\nParameter 2: idBeta\n",
		},
		{
			name:           "overlapping route - specific versioned path with alphanumeric params",
			path:           "/api/v3/user123/data456/version",
			expectedStatus: http.StatusOK,
			expectedBody:   "Path parameters received:\nParameter 1: user123\nParameter 2: data456\n",
		},
		{
			name:           "overlapping route - non-versioned path with alphanumeric params",
			path:           "/api/v3/user123/data456",
			expectedStatus: http.StatusOK,
			expectedBody:   "Path parameters received:\nParameter 1: user123\nParameter 2: data456\n",
		},
		{
			name:           "overlapping route - general path with alphanumeric params",
			path:           "/api/v3/item789/detail000",
			expectedStatus: http.StatusOK,
			expectedBody:   "Path parameters received:\nParameter 1: item789\nParameter 2: detail000\n",
		},
		{
			name:           "overlapping route - path does not match general or specific versioned",
			path:           "/api/v3/id1/id2/otherextension",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "overlapping route - path too short for general",
			path:           "/api/v3/id1",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "overlapping route - path too long for general and not matching specific",
			path:           "/api/v3/id1/id2/id3/version", // too many params for /api/v3/%s/%s/version
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "overlapping route - path too long for general",
			path:           "/api/v3/id1/id2/id3", // too many params for /api/v3/%s/%s
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "additional path after /api/v3/id1/id2/version",
			method:         http.MethodGet,
			path:           "/api/v3/id1/id2/version/extra",
			expectedStatus: http.StatusNotFound,
			expectedBody:   "404 page not found\n",
		},
		{
			name:           "invalid method for a valid custom route path (/api/v3/.../version)",
			method:         http.MethodPost,
			path:           "/api/v3/id1/id2/version",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
		{
			name:           "invalid method for a valid custom route path (/api/v3/...)",
			method:         http.MethodPut,
			path:           "/api/v3/idAlpha/idBeta",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			requestMethod := tt.method
			if requestMethod == "" {
				requestMethod = http.MethodGet
			}

			req, err := http.NewRequest(requestMethod, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, tt.expectedStatus)
			}

			// Trim newlines for consistent comparison, as http.Error and fmt.Fprintln might differ in trailing newlines.
			if strings.TrimSpace(rr.Body.String()) != strings.TrimSpace(tt.expectedBody) {
				t.Errorf("handler returned unexpected body: got %q want %q",
					rr.Body.String(), tt.expectedBody)
			}
		})
	}
}
