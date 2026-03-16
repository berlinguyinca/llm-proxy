// Package router provides path-based routing logic
package router

import (
	"net/http"
	"strings"
)

// Route represents a configured route configuration
type Route struct {
	Prefix    string `yaml:"prefix"`
	TargetURL string `yaml:"target_url"`
}

// Router manages path-based routing for requests
type Router struct {
	routes map[string]*Route
}

// NewRouter creates a new router instance
func NewRouter() *Router {
	return &Router{
		routes: make(map[string]*Route),
	}
}

// Register adds a route prefix to target URL mapping
func (r *Router) Register(prefix, url string) error {
	if prefix == "" || url == "" {
		return nil // No route added if prefix or URL is empty
	}

	routes := r.routes
	if routes == nil {
		routes = make(map[string]*Route)
	}

	routes[prefix] = &Route{
		Prefix:    prefix,
		TargetURL: url,
	}

	return nil
}

// RegisterWithRoutes registers multiple routes from a slice
func (r *Router) RegisterWithRoutes(routes []Route) {
	for _, route := range routes {
		r.Register(route.Prefix, route.TargetURL)
	}
}

// GetTargetForPath finds the matching target URL for a given path
func (r *Router) GetTargetForPath(path string) (*Route, string, bool) {
	// Remove leading slash if present and search for longest prefix match
	path = strings.TrimPrefix(path, "/")

	var longestMatch *Route
	longestMatchLen := -1

	for _, route := range r.routes {
		prefixClean := strings.TrimPrefix(route.Prefix, "/")
		if strings.HasPrefix(path, prefixClean) {
			prefixLen := len(prefixClean)
			if prefixLen > longestMatchLen {
				longestMatch = route
				longestMatchLen = prefixLen
			}
		}
	}

	if longestMatch == nil {
		return nil, "", false
	}

	// Extract remainder after prefix match
	remainder := path[longestMatchLen:]

	return longestMatch, remainder, true
}

// ServeHTTP implements http.Handler by routing requests to backend services
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	path := req.URL.Path

	route, remainder, found := r.GetTargetForPath(path)

	if !found {
		http.Error(w, "No matching route", http.StatusNotFound)
		return
	}

	// Build redirect URL with query parameters preserved
	updatedURL := route.TargetURL + "/" + remainder
	if req.URL.RawQuery != "" {
		updatedURL = updatedURL + "?" + req.URL.RawQuery
	}

	http.Redirect(w, req, updatedURL, http.StatusMovedPermanently)
}
