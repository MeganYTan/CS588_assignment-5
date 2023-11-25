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
    _ "github.com/lib/pq"
)

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





// getLastNDayGitHubIssues fetches issues related to a topic created in the last N days.
func getLastNDayGitHubIssues(topic string, days int) ([]GitHubIssue, error) {
    since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
    url := fmt.Sprintf("https://api.github.com/search/issues?q=%s+type:issue+created:>=%s", topic, since)

    apiToken := "your_github_api_token"
    req.Header.Set("Authorization", "token " + apiToken)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Println("error making the request: %v", err)
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println("error reading the response: %v", err)
        return nil, err
    }

    var searchResult SearchResult
    err = json.Unmarshal(body, &searchResult)
    if err != nil {
        log.Println("error decoding JSON: %v", err)
        return nil, err
    }

    return searchResult.Items, nil
}

// Function to insert issues into the database
func insertIssues(db *sql.DB, issues []GitHubIssue, days int) error {
    tableName := fmt.Sprintf("github_issues_%d", days)
    // Drop the table if it exists
    dropTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)
    _, err := db.Exec(dropTableSQL)
    if err != nil {
        log.Printf("Error dropping table %s: %v", tableName, err)
        return err
    }

    // Create the table
    createTableSQL := fmt.Sprintf(`
    CREATE TABLE %s (
        id SERIAL PRIMARY KEY,
        issue_id INT UNIQUE NOT NULL,
        title TEXT NOT NULL,
        body TEXT,
        created_at TIMESTAMP NOT NULL,
        updated_at TIMESTAMP NOT NULL
    );`, tableName)
    _, err = db.Exec(createTableSQL)
    if err != nil {
        log.Printf("Error creating table %s: %v", tableName, err)
        return err
    }

    for _, issue := range issues {
        _, err := db.Exec(fmt.Sprintf("INSERT INTO %s (issue_id, title, body, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)", tableName),
            issue.ID, issue.Title, issue.Body, issue.CreatedAt, issue.UpdatedAt)
        if err != nil {
            return err
        }
    }
    return nil
}

func fetchAndStoreIssues(db *sql.DB, topic string, days int) error {
    issues, err := getLastNDayGitHubIssues(topic, days)
    if err != nil {
        log.Println("error fetching issues for %s: %v", topic, err)
        return err
    }

    if len(issues) == 0 {
        log.Println("No new issues found for %s\n", topic)
        return nil
    }

    err = insertIssues(db, issues, days)
    if err != nil {
        log.Println("error inserting issues for %s into database: %v", topic, err)
        return err
    }

    log.Println("Successfully inserted %d issues for %s into the database.\n", len(issues), topic)
    return nil
}

func getRepoLastNDaysGitHubIssues(repo string, days int) ([]GitHubIssue, error) {
    since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
    url := fmt.Sprintf("https://api.github.com/search/issues?q=repo:%s+type:issue+created:>=%s", repo, since)

    resp, err := http.Get(url)
    if err != nil {
        log.Println("error making the request: %v", err)
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println("error reading the response: %v", err)
        return nil, err
    }

    var searchResult SearchResult
    err = json.Unmarshal(body, &searchResult)
    if err != nil {
        log.Println("error decoding JSON: %v", err)
        return nil, err
    }

    return searchResult.Items, nil
}


// DB
// Function to open a connection to the PostgreSQL database
func openDB() (*sql.DB, error) {
    connectionName := "assignment-5-406009:us-central1:mypostgres"
	dbUser := "postgres"
	dbPass := "root"
	dbName := "assignment_5"

	dbURI := fmt.Sprintf("host=%s dbname=%s user=%s password=%s sslmode=disable",
		connectionName, dbName, dbUser, dbPass)

	// Initialize the SQL DB handle
	log.Println("Initializing database connection")
	db, err := sql.Open("cloudsqlpostgres", dbURI)
	if err != nil {
		log.Fatalf("Error on initializing database connection: %s", err.Error())
        return nil, err
	}
	// defer db.Close()

	//Test the database connection
	log.Println("Testing database connection")
	err = db.Ping()
	if err != nil {
		log.Fatalf("Error on database connection: %s", err.Error())
        return nil, err
	}
	log.Println("Database connection established")

	return db, nil
}
func main() {
    // db connection
    db, err := openDB()
    if err != nil {
        log.Fatalf("Error opening database: %v", err)
    }
    defer db.Close()

    // github
    topics := []string{"Selenium", "Docker", "Milvus"}
    daysList := []int{2, 7, 45} // List of timeframes to check

    for _, days := range daysList {
        log.Println("Fetching issues for the past %d days\n", days)
        for _, topic := range topics {
            err := fetchAndStoreIssues(db, topic, days)
            if err != nil {
                log.Println(err)
                continue
            }
        }
    }

    repos := []string{"prometheus/prometheus", "golang/go"}
    for _, days := range daysList {
        log.Println("Fetching issues for the past %d days\n", days)
        for _, repo := range repos {
            issues, err := getRepoLastNDaysGitHubIssues(repo, days)
            if err != nil {
                log.Println(err)
                continue
            }

            if len(issues) == 0 {
                log.Println("No new issues found for %s\n", repo)
                continue
            }

            err = insertIssues(db, issues, days)
            if err != nil {
                log.Println("Error inserting issues for %s into database: %v\n", repo, err)
                continue
            }

            log.Println("Successfully inserted %d issues for %s into the database.\n", len(issues), repo)
        }
    }
}