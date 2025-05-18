package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gorilla/mux"
)

func main() {
	// Set up a simple router to test route printing
	router := mux.NewRouter()

	// Add example routes
	router.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET")
	router.HandleFunc("/api/login", func(w http.ResponseWriter, r *http.Request) {}).Methods("POST")

	// Create a subrouter for authenticated routes
	apiRouter := router.PathPrefix("/api").Subrouter()

	// Register some example route
	apiRouter.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET")
	apiRouter.HandleFunc("/users/{id}", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET", "PUT", "DELETE")
	apiRouter.HandleFunc("/symbols", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET", "POST")
	apiRouter.HandleFunc("/symbols/{id}", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET", "PUT", "DELETE")
	apiRouter.HandleFunc("/family", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET", "POST")
	apiRouter.HandleFunc("/trades", func(w http.ResponseWriter, r *http.Request) {}).Methods("GET", "POST")

	// Add a catch-all route
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

	// Print all routes
	fmt.Println("\n=== EXAMPLE ROUTES THAT WOULD BE IN YOUR APP ===")
	fmt.Println("METHOD\tPATH")
	fmt.Println("-------------------------------")

	err := router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
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

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("==============================")

	fmt.Println("\nTo display routes in your actual app:")
	fmt.Println("1. Add the following function to your internal/api/router_utils.go file:")
	fmt.Println(`
func PrintRoutes(r *mux.Router) {
	fmt.Println("\n=== Registered Routes ===")
	fmt.Println("METHOD\tPATH")
	fmt.Println("-------------------------------")
	
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil {
			return nil
		}
		
		methods, err := route.GetMethods()
		methodStr := "ANY"
		if err == nil && len(methods) > 0 {
			methodStr = strings.Join(methods, ",")
		}
		
		fmt.Printf("%s\t%s\n", methodStr, pathTemplate)
		return nil
	})
	
	fmt.Println("==============================")
}`)
	fmt.Println("\n2. Call this function at the end of your SetupRouter function in router.go")

	os.Exit(0) // Don't start the HTTP server
}
