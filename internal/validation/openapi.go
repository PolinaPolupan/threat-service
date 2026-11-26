package validation

import (
    "context"
    "net/http"

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

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            route, pathParams, err := router.FindRoute(r)
            if err != nil {
                http.Error(w, "Invalid route", http.StatusBadRequest)
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
