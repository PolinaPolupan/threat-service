package validation

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/getkin/kin-openapi/routers/gorillamux"
)

func LoadSpec(path string) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = false

	spec, err := loader.LoadFromFile(path)
	if err != nil {
		return nil, err
	}

	if err := spec.Validate(context.Background()); err != nil {
		return nil, err
	}

	return spec, nil
}

func NewMiddleware(spec *openapi3.T) (func(http.Handler) http.Handler, error) {
	router, err := gorillamux.NewRouter(spec)
	if err != nil {
		return nil, err
	}

	if len(spec.Servers) == 0 {
		return nil, fmt.Errorf("OpenAPI spec has no servers")
	}

	serverURL, err := url.Parse(spec.Servers[0].URL)
	if err != nil {
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			routeRequest := r.Clone(r.Context())
			routeRequest.Host = serverURL.Host

			route, pathParams, err := router.FindRoute(routeRequest)
			if err != nil {
				http.Error(w, fmt.Sprintf("Invalid route: %v", err), http.StatusBadRequest)
				return
			}

			input := &openapi3filter.RequestValidationInput{
				Request:    r,
				PathParams: pathParams,
				Route:      route,
			}

			if err := openapi3filter.ValidateRequest(r.Context(), input); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			next.ServeHTTP(w, r)
		})
	}, nil
}
