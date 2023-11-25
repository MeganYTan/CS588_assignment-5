package main

import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

// Define data structures for API responses
type StackOverflowPost struct {
    Title string `json:"title"`
    Link  string `json:"link"`
}
type StackOverflowResponse struct {
    Items []StackOverflowPost `json:"items"`
}

type GitHubIssue struct {
    Title string `json:"title"`
    URL   string `json:"html_url"`
}

// Function to fetch data from StackOverflow
func fetchStackOverflowData(query string) ([]StackOverflowPost, error) {
	start := time.Now()
    defer func() {
        apiCalls.With(prometheus.Labels{"api": "stackoverflow"}).Inc()
        apiCallDuration.With(prometheus.Labels{"api": "stackoverflow"}).Observe(time.Since(start).Seconds())
    }()
    url := fmt.Sprintf("https://api.stackexchange.com/2.3/search?order=desc&sort=activity&intitle=%s&site=stackoverflow", query)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result StackOverflowResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result.Items, nil // This now returns []StackOverflowPost, matching the function's signature
}

// Function to fetch data from GitHub
func fetchGitHubIssues(repo string) ([]GitHubIssue, error) {
	start := time.Now()
    defer func() {
        apiCalls.With(prometheus.Labels{"api": "github"}).Inc()
        apiCallDuration.With(prometheus.Labels{"api": "github"}).Observe(time.Since(start).Seconds())
    }()
    url := fmt.Sprintf("https://api.github.com/repos/%s/issues", repo)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var issues []GitHubIssue
    if err := json.NewDecoder(resp.Body).Decode(&issues); err != nil {
		return nil, err
    }

    return issues, nil
}


// Prometheus metrics
var (
    apiCalls = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "api_calls_total",
            Help: "Total number of API calls made.",
        },
        []string{"api"},
    )

    apiCallDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "api_call_duration_seconds",
            Help:    "Duration of API calls.",
            Buckets: prometheus.DefBuckets,
        },
        []string{"api"},
    )
)

func init() {
    // Register Prometheus metrics
    prometheus.MustRegister(apiCalls)
    prometheus.MustRegister(apiCallDuration)
}


func main() {
    // Set up a HTTP server for Prometheus metrics
    http.Handle("/metrics", promhttp.Handler())
    go func() {
        fmt.Println("Serving Prometheus metrics on :9090")
		if err := http.ListenAndServe(":9090", nil); err != nil {
            fmt.Println("Error starting Prometheus metrics server:", err)
        }
    }()

    // Example usage - make sure these lines are present and not commented
    fmt.Println("Fetching data from StackOverflow and GitHub...")
    soData, err := fetchStackOverflowData("Prometheus")
    if err != nil {
        fmt.Println("Error fetching StackOverflow data:", err)
        return
    }
    fmt.Println("StackOverflow Data:", soData)

    ghIssues, err := fetchGitHubIssues("prometheus/prometheus")
    if err != nil {
        fmt.Println("Error fetching GitHub issues:", err)
        return
    }
    fmt.Println("GitHub Issues:", ghIssues)
	select {}
}
// func main() {
//     // Set up a HTTP server for Prometheus metrics in a separate goroutine
//     http.Handle("/metrics", promhttp.Handler())
//     go func() {
//         fmt.Println("Serving Prometheus metrics on :9090")
//         if err := http.ListenAndServe(":9090", nil); err != nil {
//             fmt.Println("Error starting Prometheus metrics server:", err)
//         }
//     }()

//     // Block the main goroutine to keep the server running
//     select {}
// }
