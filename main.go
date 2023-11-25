package main
// go run main.go to run this
import (
    "encoding/json"
    "fmt"
    "net/http"
    "time"
    "io/ioutil"
    "log"
	"database/sql"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promhttp"
    _ "github.com/lib/pq"
)

// Define data structures for API responses
type StackOverflowPost struct {
    Title string `json:"title"`
    Link  string `json:"link"`
}
type StackOverflowResponse struct {
    Items []StackOverflowPost `json:"items"`
}

// GitHubIssue represents a single GitHub issue.
type GitHubIssue struct {
    URL           string    `json:"url"`
    RepositoryURL string    `json:"repository_url"`
    LabelsURL     string    `json:"labels_url"`
    CommentsURL   string    `json:"comments_url"`
    EventsURL     string    `json:"events_url"`
    HTMLURL       string    `json:"html_url"`
    ID            int       `json:"id"`
    NodeID        string    `json:"node_id"`
    Number        int       `json:"number"`
    Title         string    `json:"title"`
    User          User      `json:"user"`
    Labels        []Label   `json:"labels"`
    State         string    `json:"state"`
    Locked        bool      `json:"locked"`
    Assignee      User      `json:"assignee"`
    Assignees     []User    `json:"assignees"`
    Comments      int       `json:"comments"`
    CreatedAt     time.Time `json:"created_at"`
    UpdatedAt     time.Time `json:"updated_at"`
    ClosedAt      time.Time `json:"closed_at"`
    Body          string    `json:"body"`
}

// User represents a GitHub user.
type User struct {
    Login             string `json:"login"`
    ID                int    `json:"id"`
    NodeID            string `json:"node_id"`
    AvatarURL         string `json:"avatar_url"`
    GravatarID        string `json:"gravatar_id"`
    URL               string `json:"url"`
    HTMLURL           string `json:"html_url"`
    FollowersURL      string `json:"followers_url"`
    FollowingURL      string `json:"following_url"`
    GistsURL          string `json:"gists_url"`
    StarredURL        string `json:"starred_url"`
    SubscriptionsURL  string `json:"subscriptions_url"`
    OrganizationsURL  string `json:"organizations_url"`
    ReposURL          string `json:"repos_url"`
    EventsURL         string `json:"events_url"`
    ReceivedEventsURL string `json:"received_events_url"`
    Type              string `json:"type"`
    SiteAdmin         bool   `json:"site_admin"`
}

// Label represents a label assigned to an issue.
type Label struct {
    ID      int    `json:"id"`
    NodeID  string `json:"node_id"`
    URL     string `json:"url"`
    Name    string `json:"name"`
    Color   string `json:"color"`
    Default bool   `json:"default"`
}

// SearchResult represents the GitHub API search result.
type SearchResult struct {
    TotalCount int            `json:"total_count"`
    Items      []GitHubIssue  `json:"items"`
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

// getLastThreeDaysGitHubIssues fetches issues related to a topic created in the last 3 days.
func getLastThreeDaysGitHubIssues(topic string) ([]GitHubIssue, error) {
    threeDaysAgo := time.Now().AddDate(0, 0, -3).Format("2006-01-02")
    url := fmt.Sprintf("https://api.github.com/search/issues?q=%s+type:issue+created:>=%s", topic, threeDaysAgo)

    resp, err := http.Get(url)
    if err != nil {
        return nil, fmt.Errorf("error making the request: %v", err)
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("error reading the response: %v", err)
    }

    var searchResult SearchResult
    err = json.Unmarshal(body, &searchResult)
    if err != nil {
        return nil, fmt.Errorf("error decoding JSON: %v", err)
    }

    return searchResult.Items, nil
}

// Function to insert issues into the database
func insertIssues(db *sql.DB, issues []GitHubIssue) error {
    for _, issue := range issues {
        _, err := db.Exec("INSERT INTO github_issues (issue_id, title, body, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)",
            issue.ID, issue.Title, issue.Body, issue.CreatedAt, issue.UpdatedAt)
        if err != nil {
            return err
        }
    }
    return nil
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

// db
// Function to open a connection to the PostgreSQL database
func openDB() () {
    connectionName := "assignment-5-406009:us-central1:mypostgres"
	dbUser := "postgres"
	dbPass := "root"
	dbName := "chicago_business_intelligence"

	dbURI := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable",
		connectionName, dbName, dbUser, dbPass)

	// Initialize the SQL DB handle
	log.Println("Initializing database connection")
	db, err := sql.Open("cloudsqlpostgres", dbURI)
	if err != nil {
		log.Fatalf("Error on initializing database connection: %s", err.Error())
	}
	defer db.Close()

	//Test the database connection
	log.Println("Testing database connection")
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error on database connection: %s", err.Error())
	}
	log.Println("Database connection established")

	log.Println("Database query done!")
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

    // db connection

    // Example usage - make sure these lines are present and not commented
    // fmt.Println("Fetching data from StackOverflow and GitHub...")
    // soData, err := fetchStackOverflowData("Prometheus")
    // if err != nil {
    //     fmt.Println("Error fetching StackOverflow data:", err)
    //     return
    // }
    // fmt.Println("StackOverflow Data:", soData)

    // ghIssues, err := fetchGitHubIssues("prometheus/prometheus")
    // if err != nil {
    //     fmt.Println("Error fetching GitHub issues:", err)
    //     return
    // }
    // fmt.Println("GitHub Issues:", ghIssues)
	select {}

    // topic := "OpenAI"
    // issues, err := getLastThreeDaysGitHubIssues(topic)
    // if err != nil {
    //     fmt.Println(err)
    //     return
    // }

    // fmt.Printf("Found %d issues:\n", len(issues))
    // for _, issue := range issues {
    //     fmt.Printf("#%d: %s\n", issue.Number, issue.Title)
    // }
}