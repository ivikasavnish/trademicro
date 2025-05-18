#!/bin/zsh

# Print instructions for manually adding route printing

cat << 'EOF'
To print all routes in your application, follow these steps:

1. Add this utility function to your internal/api/router_utils.go file:

```go
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
```

2. Add a call to this function at the end of your SetupRouter function in router.go:

```go
// At the end of your SetupRouter function, right before returning the router
PrintRoutes(router)
return router
```

3. You can also add a special endpoint to view routes:

```go
// Add this with your other routes
router.HandleFunc("/api/debug/routes", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "text/plain")
    
    fmt.Fprintln(w, "\n=== Registered Routes ===")
    fmt.Fprintln(w, "METHOD\tPATH")
    fmt.Fprintln(w, "-------------------------------")
    
    router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
        pathTemplate, err := route.GetPathTemplate()
        if err != nil {
            return nil
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
}).Methods("GET")
```

4. Then when you run your app, you'll see all routes printed to the console, or you can visit /api/debug/routes to see them in your browser.
EOF
