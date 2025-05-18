package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// PrintRoutes walks through all routes registered in the router and prints them
func PrintRoutes(r *mux.Router) {
	fmt.Println("\n=== Registered Routes ===")
	fmt.Println("METHOD\tPATH\tHANDLER")
	fmt.Println("-------------------------------")

	_ = r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, _ := route.GetPathTemplate()
		methods, _ := route.GetMethods()

		// If no methods are specified, assume all methods
		methodStr := "ANY"
		if len(methods) > 0 {
			methodStr = strings.Join(methods, ",")
		}

		// Get handler name where possible
		handlerName := "-"
		if route.GetHandler() != nil {
			// Attempt to get function name, but this is complex
			// and might not always work as expected
			handlerName = fmt.Sprintf("%T", route.GetHandler())
		}

		fmt.Printf("%s\t%s\t%s\n", methodStr, pathTemplate, handlerName)
		return nil
	})
	fmt.Println("==============================\n")
}

// PrintRoutesHandler returns a handler function to print all routes
func PrintRoutesHandler(router *mux.Router) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		fmt.Fprintln(w, "=== Registered Routes ===")
		fmt.Fprintln(w, "METHOD\tPATH\tHANDLER")
		fmt.Fprintln(w, "-------------------------------")

		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			pathTemplate, err := route.GetPathTemplate()
			if err != nil {
				pathTemplate = "<unknown>"
			}

			methods, err := route.GetMethods()
			methodStr := "ANY"
			if err == nil && len(methods) > 0 {
				methodStr = strings.Join(methods, ",")
			}

			fmt.Fprintf(w, "%s\t%s\n", methodStr, pathTemplate)
			return nil
		})

		fmt.Fprintln(w, "==============================")
	}
}
