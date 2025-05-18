package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	// Create a new router
	router := mux.NewRouter()

	// Add some example routes
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET")
	router.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {}).Methods("POST")

	// Create a subrouter
	apiRouter := router.PathPrefix("/api").Subrouter()
	authRouter := apiRouter.PathPrefix("").Subrouter()

	// Add routes to the subrouter
	authRouter.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET")
	authRouter.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET")
	authRouter.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {}).Methods("POST")

	// Add a route with multiple methods
	authRouter.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET", "POST")

	// Print all routes
	fmt.Println("\n=== Registered Routes ===")
	fmt.Println("METHOD\tPATH")
	fmt.Println("-------------------------------")

	err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil {
			pathTemplate = "<unknown>"
		}

		methods, err := route.GetMethods()
		methodStr := "ANY"
		if err == nil && len(methods) > 0 {
			methodStr = ""
			for i, method := range methods {
				if i > 0 {
					methodStr += ","
				}
				methodStr += method
			}
		}

		fmt.Printf("%s\t%s\n", methodStr, pathTemplate)
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("==============================")

	fmt.Println("\nTo use this with your actual router in your app, add similar walking code to your main.go")
	fmt.Println("Or run your app with a route that prints all routes")

	os.Exit(0) // Exit without starting the server
}
