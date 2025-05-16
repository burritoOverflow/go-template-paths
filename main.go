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

func dynamicPathHandler(w http.ResponseWriter, r *http.Request) {
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
func dynamicPathHandlerFunc(w http.ResponseWriter, r *http.Request) {
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
		// Method 2: Using http.ServeMux with regexp in the handler
		mux := http.NewServeMux()
		// Match the prefix and parse inside
		mux.HandleFunc("/foo/bar/", dynamicPathHandler)
		log.Printf("Starting server with ServeMux on %s", addr)
		log.Fatal(http.ListenAndServe(addr, mux))
	}
}
