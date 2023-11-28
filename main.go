package main

import (
	"fmt"
    "io/ioutil"
	"net/http"
	"time"
	"os"
	"log"
	"database/sql"
	"encoding/json"
	_ "github.com/lib/pq"
	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
)
const baseURL string = "https://api.stackexchange.com/2.3"
const stackApiKey string = "5tYPJGJ2XmHpHwDrZZ51nA(("
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

// StackExchanges structs
type Question struct {
	QuestionID   int    `json:"question_id"`
	Title        string `json:"title"`
	CreationDate int64  `json:"creation_date"`
}

type Answer struct {
	AnswerID     int    `json:"answer_id"`
	QuestionID   int    `json:"question_id"`
	CreationDate int64  `json:"creation_date"`
}

type StackExchangeResponse struct {
	Items []Question `json:"items"`
}
type StackExchangeAnswerResponse struct {
	Items []Answer `json:"items"`
}
func main() {

	// Database connection settings
	// connectionName := "pivotal-data-406222:us-central1:mypostgres"
	// dbUser := "postgres"
	// dbPass := "root"
	// dbName := "assignment-5"

	connectionName := "assignment-5-406009:us-central1:mypostgres"
	dbUser := "postgres"
	dbPass := "root"
	dbName := "assignment-5"

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

	port := os.Getenv("PORT")
	if port == "" {
        port = "8080"
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello, world!"))
    })
	go func() {
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
	}()
	
    // // github
    topics := []string{"Selenium", "Docker", "Milvus"}
    daysList := []int{2, 7, 45} // List of timeframes to check

    for _, days := range daysList {
        log.Printf("Fetching issues for the past %d days\n", days)
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
        log.Printf("Fetching issues for the past %d days\n", days)
        for _, repo := range repos {
            issues, err := getRepoLastNDaysGitHubIssues(repo, days)
            if err != nil {
                log.Println(err)
                continue
            }

            if len(issues) == 0 {
                log.Printf("No new issues found for %s\n", repo)
                continue
            }

            err = insertIssues(db, issues, days, repo)
            if err != nil {
                log.Printf("Error inserting issues for %s into database: %v\n", repo, err)
                continue
            }

            log.Printf("Successfully inserted %d issues for %s into the database.\n", len(issues), repo)
        }
    }

	// stack overflow
	tags := []string{"Selenium", "Docker", "Milvus", "Prometheus", "Go"}
    for _, days := range daysList {
        log.Printf("Fetching issues for the past %d days\n", days)
        for _, tag := range tags {
            err := fetchAndStoreStackOverflowData(db, tag, days)
            if err != nil {
                log.Println(err)
                continue
            }
        }
    }
    // keeps server up
	for {
		log.Println("Inside For")
		time.Sleep(24 * time.Hour)
	}
}
// stack overflow
func getStackOverflowQuestionsLastNDays(tag string, days int) ([]Question, error) {
    today := time.Now().Unix()
    since := time.Now().AddDate(0, 0, -days).Unix()
    url := fmt.Sprintf("%s/questions?fromdate=%d&todate=%d&order=desc&sort=activity&tagged=%s&site=stackoverflow&key=%s", baseURL, since, today, tag, stackApiKey)

    resp, err := http.Get(url)
    if err != nil {
        log.Printf("error making the stackoverflow question api request: %v", err)
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Printf("error reading the stackoverflow question api response: %v", err)
        return nil, err
    }

    var searchResult StackExchangeResponse
    err = json.Unmarshal(body, &searchResult)
    if err != nil {
        log.Printf("error decoding stackoverflow question api JSON: %v", err)
        return nil, err
    }

    return searchResult.Items, nil
}
func getStackOverflowAnswersLastNDaysForQuestion(questionIds []int, days int) ([]Answer, error) {
    today := time.Now().Unix()
    since := time.Now().AddDate(0, 0, -days).Unix()
    var answers []Answer

    for _, id := range questionIds {
    // /2.3/questions/{ids}/answers?fromdate=%d&todate=%d&order=desc&sort=activity&site=stackoverflow&key=%s
        url := fmt.Sprintf("%s/questions/%d/answers?fromdate=%d&todate=%d&order=desc&sort=activity&site=stackoverflow&key=%s", baseURL, id, since, today, stackApiKey)

        resp, err := http.Get(url)
        if err != nil {
            log.Printf("error making the stackoverflow answer api request: %v", err)
            return nil, err
        }
        defer resp.Body.Close()

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            log.Printf("error reading the stackoverflow answer api response: %v", err)
            return nil, err
        }

        var searchResult StackExchangeAnswerResponse
        err = json.Unmarshal(body, &searchResult)
        answers = append(answers, searchResult.Items...)
        if err != nil {
            log.Printf("error decoding stackoverflow answer api JSON: %v", err)
            return nil, err
        }
    }
    return answers, nil
}

func insertStackoverflowQuestions(db *sql.DB, data []Question, days int, tag string) error {
    tableName := fmt.Sprintf("stackoverflow_questions_%d", days)
    dropTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)
    _, err := db.Exec(dropTableSQL)
    if err != nil {
        log.Printf("Error dropping table %s: %v", tableName, err)
        return err
    }

    // Create the table
    createTableSQL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id SERIAL PRIMARY KEY,
        owner_id INTEGER NOT NULL,
		post_id INTEGER NOT NULL,
        tag TEXT,
		title TEXT,
		creation_date TEXT
	)`, tableName)
    _, err = db.Exec(createTableSQL)
    if err != nil {
        log.Printf("Error creating table %s: %v", tableName, err)
        return err
    }


	for _, item := range data {
		query := fmt.Sprintf(`INSERT INTO %s (post_id, title, creation_date, tag) VALUES ($1, $2, $3, $4)`, tableName)
		_, err := db.Exec(query, item.QuestionID, item.Title, item.CreationDate, tag)
		if err != nil {
			log.Println("Error inserting stackoverflow questions data: ", err)
		}
	}
    return nil
}

func insertStackoverflowAnswers(db *sql.DB, data []Answer, days int, tag string) error {
    tableName := fmt.Sprintf("stackoverflow_answers_%d", days)
    dropTableSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s;", tableName)
    _, err := db.Exec(dropTableSQL)
    if err != nil {
        log.Printf("Error dropping table %s: %v", tableName, err)
        return err
    }

    // Create the table
    createTableSQL := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s (
		id SERIAL PRIMARY KEY,
		post_id INTEGER NOT NULL,
        tag TEXT,
		question_id INTEGER,
		creation_date TEXT
	)`, tableName)
    _, err = db.Exec(createTableSQL)
    if err != nil {
        log.Printf("Error creating table %s: %v", tableName, err)
        return err
    }


	for _, item := range data {
		query := fmt.Sprintf(`INSERT INTO %s (post_id, question_id, creation_date, tag) VALUES ($1, $2, $3, $4)`, tableName)
		_, err := db.Exec(query, item.AnswerID, item.QuestionID, item.CreationDate, tag)
		if err != nil {
			log.Println("Error inserting stackoverflow questions data: ", err)
		}
	}
    return nil
}

func fetchAndStoreStackOverflowData(db *sql.DB, topic string, days int) error {
    questions, err := getStackOverflowQuestionsLastNDays(topic, days)
    if err != nil {
        log.Printf("error fetching questions for %s: %v", topic, err)
        return err
    }

    if len(questions) == 0 {
        log.Printf("No new issues found for %s\n", topic)
        return nil
    }

    err = insertStackoverflowQuestions(db, questions, days, topic)
    if err != nil {
        log.Printf("error inserting questions for %s into database: %v", topic, err)
        return err
    }

    log.Printf("Successfully inserted %d questions for %s into the stackoverflow database.\n", len(questions), topic)

    // answers
    var questionIds []int
    for _, q := range questions {
        questionIds = append(questionIds, q.QuestionID)
    }

    // func getStackOverflowAnswersLastNDaysForQuestion(questionIds string, days int) ([]map[string]interface{}, error)
    answers, err := getStackOverflowAnswersLastNDaysForQuestion(questionIds, days)
    if err != nil {
        fmt.Println("Error fetching answers:", err)
    }
    //func insertStackoverflowAnswers(db *sql.DB, data []Answer, days int, tag string) error {
    err = insertStackoverflowAnswers(db, answers, days, topic)
    if err != nil {
        log.Printf("error inserting answers for %s into database: %v", topic, err)
        return err
    }

    log.Printf("Successfully inserted %d answers for %s into the stackoverflow database.\n", len(questions), topic)
    
    return nil
    
}


// getLastNDayGitHubIssues fetches issues related to a topic created in the last N days.
func getLastNDayGitHubIssues(topic string, days int) ([]GitHubIssue, error) {
    since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
    url := fmt.Sprintf("https://api.github.com/search/issues?q=%s+type:issue+created:>=%s", topic, since)
    req, err := http.NewRequest("GET", url, nil)
    if err != nil {
        log.Printf("error creating request: %v", err)
        return nil, err
    }
    apiToken := "your_github_api_token"
    req.Header.Set("Authorization", "token " + apiToken)
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        log.Printf("error making the request: %v", err)
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Printf("error reading the response: %v", err)
        return nil, err
    }

    var searchResult SearchResult
    err = json.Unmarshal(body, &searchResult)
    if err != nil {
        log.Printf("error decoding JSON: %v", err)
        return nil, err
    }

    return searchResult.Items, nil
}

// Function to insert issues into the database
func insertIssues(db *sql.DB, issues []GitHubIssue, days int, topic string) error {
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
        issue_number INT,
        topic TEXT,
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
        _, err := db.Exec(fmt.Sprintf("INSERT INTO %s (issue_id, issue_number, user_id, topic, title, body, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)", tableName),
            issue.ID, issue.Number, issue.User.ID, topic, issue.Title, issue.Body, issue.CreatedAt, issue.UpdatedAt)
        if err != nil {
            return err
        }
    }
    return nil
}

func fetchAndStoreIssues(db *sql.DB, topic string, days int) error {
    issues, err := getLastNDayGitHubIssues(topic, days)
    if err != nil {
        log.Printf("error fetching issues for %s: %v", topic, err)
        return err
    }

    if len(issues) == 0 {
        log.Printf("No new issues found for %s\n", topic)
        return nil
    }

    err = insertIssues(db, issues, days, topic)
    if err != nil {
        log.Printf("error inserting issues for %s into database: %v", topic, err)
        return err
    }

    log.Printf("Successfully inserted %d issues for %s into the github database.\n", len(issues), topic)

    // then do repo specific
    return nil
}

func getRepoLastNDaysGitHubIssues(repo string, days int) ([]GitHubIssue, error) {
    since := time.Now().AddDate(0, 0, -days).Format("2006-01-02")
    url := fmt.Sprintf("https://api.github.com/search/issues?q=repo:%s+type:issue+created:>=%s", repo, since)

    resp, err := http.Get(url)
    if err != nil {
        log.Printf("error making the request: %v", err)
        return nil, err
    }
    defer resp.Body.Close()

    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Printf("error reading the response: %v", err)
        return nil, err
    }

    var searchResult SearchResult
    err = json.Unmarshal(body, &searchResult)
    if err != nil {
        log.Printf("error decoding JSON: %v", err)
        return nil, err
    }

    return searchResult.Items, nil
}