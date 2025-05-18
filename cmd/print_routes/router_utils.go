package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

// PrintRoutes walks through all routes registered in the router and prints them
func PrintRoutes(r *mux.Router) {
	fmt.Println("\n=== Registered Routes ===")
	fmt.Println("METHOD\tPATH")
	fmt.Println("-------------------------------")

	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil {
			return nil // Skip routes without templates
		}

		methods, err := route.GetMethods()
		methodStr := "ANY"
		if err == nil && len(methods) > 0 {
			methodStr = strings.Join(methods, ",")
		}

		fmt.Printf("%s\t%s\n", methodStr, pathTemplate)
		return nil
	})

	fmt.Println("==============================\n")
}

// PrintRoutesHandler returns a handler function that prints all routes to the response
func PrintRoutesHandler(router *mux.Router) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		fmt.Fprintln(w, "\n=== Registered Routes ===")
		fmt.Fprintln(w, "METHOD\tPATH")
		fmt.Fprintln(w, "-------------------------------")

		router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
			pathTemplate, err := route.GetPathTemplate()
			if err != nil {
				return nil // Skip routes without templates
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
