package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strings"
)

// simple hardcoded implementation
func dynamicPathHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Define the pattern we're looking for
	pathPattern := regexp.MustCompile(`^/foo/bar/([a-zA-Z0-9]+)/baz/([a-zA-Z0-9]+)/qux$`)
	matches := pathPattern.FindStringSubmatch(r.URL.Path)
	if matches == nil {
		log.Printf("No matches for pattern %s in path '%s'", pathPattern, r.URL.Path)
		http.NotFound(w, r)
		return
	}

	// Extract the two path parameters
	log.Printf("%d Matches found: %v\n", len(matches), matches)

	param1 := matches[1]
	param2 := matches[2]

	// Well just use the parameters in the response
	fmt.Fprintf(w, "Path parameters received:\n")
	fmt.Fprintf(w, "First parameter: %s\n", param1)
	fmt.Fprintf(w, "Second parameter: %s\n", param2)
}

// Convert a provided pattern path pattern from i.e "/foo/bar/%s/baz/%s/qux" to a proper alphanumeric regex
func makeRegexPatternStr(pattern string) string {
	return "^" + strings.Replace(pattern, "%s", "([a-zA-Z0-9]+)", -1) + "$"
}

// creates an http.HandlerFunc that matches the request path against the provided templated path and extracts parameters
// i.e provide "/foo/bar/%s/baz/%s/qux" and it will match paths like "/foo/bar/123/baz/456/qux"
func newPathRegexHandler(routeTemplateStr string) http.HandlerFunc {
	// internally generate the regex pattern from the template
	regexPatternStr := makeRegexPatternStr(routeTemplateStr)
	pathPattern := regexp.MustCompile(regexPatternStr)
	numGroups := pathPattern.NumSubexp()

	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		matches := pathPattern.FindStringSubmatch(r.URL.Path)
		// The handler must ensure the *full* path matches the specific regex.
		if matches == nil {
			log.Printf("No matches for pattern '%s' in path '%s'", regexPatternStr, r.URL.Path)
			http.NotFound(w, r)
			return
		}

		// matches[0] is the full string, we're only interested in the capturing groups
		if len(matches)-1 != numGroups {
			log.Printf("Error: Expected %d capturing groups, got %d from path '%s' with pattern '%s'",
				numGroups, len(matches)-1, r.URL.Path, regexPatternStr)

			http.Error(w, "Internal server error: Mismatched capturing groups", http.StatusInternalServerError)
			return
		}

		log.Printf("%d path parameters captured from path '%s' using pattern '%s': %v (full match: %s)\n",
			numGroups, r.URL.Path, regexPatternStr, matches[1:], matches[0])

		if numGroups > 0 {
			for i := 0; i < numGroups; i++ {
				// we'll return each parameter in the response
				fmt.Fprintf(w, "Parameter %d: %s\n", i+1, matches[i+1])
			}
		} else {
			fmt.Fprintln(w, "No parameters captured.")
		}
	}
}

// associates a pattern with a handler
type route struct {
	pattern *regexp.Regexp   // compiled regex pattern matching a path, i.e "/foo/bar/%s/baz/%s/qux"
	handler http.HandlerFunc // handler function to call when the pattern matches
}

type customRouter struct {
	routes []*route
}

// register a new route with a template pattern and handler
func (r *customRouter) HandleFunc(pattern string, handler http.HandlerFunc) {
	// Convert the pattern from "/foo/bar/%s/baz/%s/qux" to a proper alphanumeric regex
	replacedRoute := strings.Replace(pattern, "%s", "([a-zA-Z0-9]+)", -1)
	log.Printf("Registering route: %s\n", replacedRoute)
	fullPattern := regexp.MustCompile("^" + replacedRoute + "$")
	r.routes = append(r.routes, &route{
		pattern: fullPattern,
		handler: handler,
	})
}

type paramKey int

func (r *customRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	for _, route := range r.routes {
		matches := route.pattern.FindStringSubmatch(req.URL.Path)
		if matches != nil {
			// Store the path parameters in the request context
			ctx := req.Context()
			// first match is the full match, ignore it
			for i, match := range matches[1:] {
				// Using the context to store params isn't ideal in plain stdlib,
				// so here we're just attaching them to the request via a custom method
				ctx = context.WithValue(ctx, paramKey(i+1), match) // Update ctx in each iteration
			}

			req = req.WithContext(ctx) // Update req once with the final context
			route.handler(w, req)
			return
		}
	}
	http.NotFound(w, req)
}

// Function to get the stored path parameters from the context
func getParam(r *http.Request, index int) string {
	value := r.Context().Value(paramKey(index))
	if value == nil {
		return ""
	}
	return value.(string)
}

// Handler function for the custom router
// TODO extend in the same manner as the mux implementation
func dynamicPathHandlerFunc(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	param1 := getParam(r, 1)
	param2 := getParam(r, 2)

	// we'll just demo return the content
	fmt.Fprintf(w, "Path parameters received:\n")
	fmt.Fprintf(w, "First parameter: %s\n", param1)
	fmt.Fprintf(w, "Second parameter: %s\n", param2)
}

func main() {
	useCustomRouter := flag.Bool("customRouter", false, "Use the custom router implementation (mux default)")
	port := flag.Int("port", 8080, "Port to run the server on")
	flag.Parse()

	addr := fmt.Sprintf(":%d", *port)

	if *useCustomRouter {
		// Method 1: Using custom router implementation
		router := &customRouter{}
		// 'populate' the template string and associate it with the handler func
		routeTemplateStr := "/foo/bar/%s/baz/%s/qux"
		router.HandleFunc(routeTemplateStr, dynamicPathHandlerFunc)
		log.Printf("Starting server with custom router on %s...", addr)
		log.Fatal(http.ListenAndServe(addr, router))
	} else {
		// Method 2: Using http.ServeMux with a generalized regex handler
		mux := http.NewServeMux()

		// Define the path template(s)
		routeTemplateStr := "/foo/bar/%s/baz/%s/qux"
		apiRouteTemplStr := "/api/v3/%s/%s"
		plainRouteTemplStr := "/blah/foo/bar/baz/qux"

		// and create and register handlers
		registerHandlerForPath(mux, routeTemplateStr)
		registerHandlerForPath(mux, apiRouteTemplStr)
		registerHandlerForPath(mux, plainRouteTemplStr)

		log.Printf("Starting server with ServeMux on %s", addr)
		log.Printf("ServeMux handles prefix '%s'", routeTemplateStr)
		log.Fatal(http.ListenAndServe(addr, mux))
	}
}

// registerHandlerForPath registers a handler for the given path template with the provided ServeMux
func registerHandlerForPath(mux *http.ServeMux, routeTemplateStr string) {
	// here, convert a path like "/foo/bar/%s/baz/%s/qux" to "/foo/bar/" to register just the 'prefix'
	// (prior to the template positions)
	pathPrefixForMux := getPathPrefix(routeTemplateStr)
	log.Printf("Registering handler for path prefix: '%s' with template %s\n", pathPrefixForMux, routeTemplateStr)

	// create and register the handler for the given path
	handler := newPathRegexHandler(routeTemplateStr)
	mux.HandleFunc(pathPrefixForMux, handler)
}

// return the prefix of the path pattern up to the first '%s' occurrence
func getPathPrefix(pattern string) string {
	// Find the first occurrence of '%s' and return the prefix up to that point
	// i.e "/foo/bar/%s/baz/%s/qux" -> "/foo/bar/"
	if idx := strings.Index(pattern, "%s"); idx != -1 {
		return pattern[:idx]
	}
	// allow for non-template routes as well
	// If no '%s' found, return the whole pattern
	return pattern
}
