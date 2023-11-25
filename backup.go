package main
 go run main.go to run this
import (
    encodingjson
    fmt
    nethttp
    time
    ioioutil
    log
	databasesql

    github.comprometheusclient_golangprometheus
    github.comprometheusclient_golangprometheuspromhttp
    _ github.comlibpq
)

 Define data structures for API responses
type StackOverflowPost struct {
    Title string `jsontitle`
    Link  string `jsonlink`
}
type StackOverflowResponse struct {
    Items []StackOverflowPost `jsonitems`
}

 GitHubIssue represents a single GitHub issue.
type GitHubIssue struct {
    URL           string    `jsonurl`
    RepositoryURL string    `jsonrepository_url`
    LabelsURL     string    `jsonlabels_url`
    CommentsURL   string    `jsoncomments_url`
    EventsURL     string    `jsonevents_url`
    HTMLURL       string    `jsonhtml_url`
    ID            int       `jsonid`
    NodeID        string    `jsonnode_id`
    Number        int       `jsonnumber`
    Title         string    `jsontitle`
    User          User      `jsonuser`
    Labels        []Label   `jsonlabels`
    State         string    `jsonstate`
    Locked        bool      `jsonlocked`
    Assignee      User      `jsonassignee`
    Assignees     []User    `jsonassignees`
    Comments      int       `jsoncomments`
    CreatedAt     time.Time `jsoncreated_at`
    UpdatedAt     time.Time `jsonupdated_at`
    ClosedAt      time.Time `jsonclosed_at`
    Body          string    `jsonbody`
}

 User represents a GitHub user.
type User struct {
    Login             string `jsonlogin`
    ID                int    `jsonid`
    NodeID            string `jsonnode_id`
    AvatarURL         string `jsonavatar_url`
    GravatarID        string `jsongravatar_id`
    URL               string `jsonurl`
    HTMLURL           string `jsonhtml_url`
    FollowersURL      string `jsonfollowers_url`
    FollowingURL      string `jsonfollowing_url`
    GistsURL          string `jsongists_url`
    StarredURL        string `jsonstarred_url`
    SubscriptionsURL  string `jsonsubscriptions_url`
    OrganizationsURL  string `jsonorganizations_url`
    ReposURL          string `jsonrepos_url`
    EventsURL         string `jsonevents_url`
    ReceivedEventsURL string `jsonreceived_events_url`
    Type              string `jsontype`
    SiteAdmin         bool   `jsonsite_admin`
}

 Label represents a label assigned to an issue.
type Label struct {
    ID      int    `jsonid`
    NodeID  string `jsonnode_id`
    URL     string `jsonurl`
    Name    string `jsonname`
    Color   string `jsoncolor`
    Default bool   `jsondefault`
}

 SearchResult represents the GitHub API search result.
type SearchResult struct {
    TotalCount int            `jsontotal_count`
    Items      []GitHubIssue  `jsonitems`
}



 Function to fetch data from StackOverflow
func fetchStackOverflowData(query string) ([]StackOverflowPost, error) {
	start = time.Now()
    defer func() {
        apiCalls.With(prometheus.Labels{api stackoverflow}).Inc()
        apiCallDuration.With(prometheus.Labels{api stackoverflow}).Observe(time.Since(start).Seconds())
    }()
    url = fmt.Sprintf(httpsapi.stackexchange.com2.3searchorder=desc&sort=activity&intitle=%s&site=stackoverflow, query)
    resp, err = http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result StackOverflowResponse
    if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return result.Items, nil  This now returns []StackOverflowPost, matching the function's signature
}


 getLastThreeDaysGitHubIssues fetches issues related to a topic created in the last 3 days.
func getLastNDayGitHubIssues(topic string, days int) ([]GitHubIssue, error) {
    since = time.Now().AddDate(0, 0, -days).Format(2006-01-02)
    url = fmt.Sprintf(httpsapi.github.comsearchissuesq=%s+typeissue+created=%s, topic, since)

    resp, err = http.Get(url)
    if err != nil {
        log.Println(error making the request %v, err)
        return nil, err
    }
    defer resp.Body.Close()

    body, err = ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println(error reading the response %v, err)
        return nil, err
    }

    var searchResult SearchResult
    err = json.Unmarshal(body, &searchResult)
    if err != nil {
        log.Println(error decoding JSON %v, err)
        return nil, err
    }

    return searchResult.Items, nil
}

 Function to insert issues into the database
func insertIssues(db sql.DB, issues []GitHubIssue, days int) error {
    tableName = fmt.Sprintf(github_issues_%d, days)
    for _, issue = range issues {
        _, err = db.Exec(fmt.Sprintf(INSERT INTO %s (issue_id, title, body, created_at, updated_at) VALUES ($1, $2, $3, $4, $5), tableName),
            issue.ID, issue.Title, issue.Body, issue.CreatedAt, issue.UpdatedAt)
        if err != nil {
            return err
        }
    }
    return nil
}

func fetchAndStoreIssues(db sql.DB, topic string, days int) error {
    issues, err = getLastNDayGitHubIssues(topic, days)
    if err != nil {
        log.Println(error fetching issues for %s %v, topic, err)
        return err
    }

    if len(issues) == 0 {
        log.Println(No new issues found for %sn, topic)
        return nil
    }

    err = insertIssues(db, issues, days)
    if err != nil {
        log.Println(error inserting issues for %s into database %v, topic, err)
        return err
    }

    log.Println(Successfully inserted %d issues for %s into the database.n, len(issues), topic)
    return nil
}

func getRepoLastNDaysGitHubIssues(repo string, days int) ([]GitHubIssue, error) {
    since = time.Now().AddDate(0, 0, -days).Format(2006-01-02)
    url = fmt.Sprintf(httpsapi.github.comsearchissuesq=repo%s+typeissue+created=%s, repo, since)

    resp, err = http.Get(url)
    if err != nil {
        log.Println(error making the request %v, err)
        return nil, err
    }
    defer resp.Body.Close()

    body, err = ioutil.ReadAll(resp.Body)
    if err != nil {
        log.Println(error reading the response %v, err)
        return nil, err
    }

    var searchResult SearchResult
    err = json.Unmarshal(body, &searchResult)
    if err != nil {
        log.Println(error decoding JSON %v, err)
        return nil, err
    }

    return searchResult.Items, nil
}


 Prometheus metrics
var (
    apiCalls = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name api_calls_total,
            Help Total number of API calls made.,
        },
        []string{api},
    )

    apiCallDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name    api_call_duration_seconds,
            Help    Duration of API calls.,
            Buckets prometheus.DefBuckets,
        },
        []string{api},
    )
)

func init() {
     Register Prometheus metrics
    prometheus.MustRegister(apiCalls)
    prometheus.MustRegister(apiCallDuration)
}

 db
 Function to open a connection to the PostgreSQL database
func openDB() (sql.DB, error) {
    connectionName = assignment-5-406009us-central1mypostgres
	dbUser = postgres
	dbPass = root
	dbName = assignment_5

	dbURI = fmt.Sprintf(host=%s dbname=%s user=%s password=%s sslmode=disable,
		connectionName, dbName, dbUser, dbPass)

	 Initialize the SQL DB handle
	log.Println(Initializing database connection)
	db, err = sql.Open(cloudsqlpostgres, dbURI)
	if err != nil {
		log.Fatalf(Error on initializing database connection %s, err.Error())
        return nil, err
	}
	 defer db.Close()

	Test the database connection
	log.Println(Testing database connection)
	err = db.Ping()
	if err != nil {
		log.Fatalf(Error on database connection %s, err.Error())
        return nil, err
	}
	log.Println(Database connection established)

	return db, nil
}
func main() {
     Set up a HTTP server for Prometheus metrics
     http.Handle(metrics, promhttp.Handler())
     go func() {
         fmt.Println(Serving Prometheus metrics on 9090)
	 	if err = http.ListenAndServe(9090, nil); err != nil {
             fmt.Println(Error starting Prometheus metrics server, err)
         }
     }()

     db connection
    db, err = openDB()
    if err != nil {
        log.Fatalf(Error opening database %v, err)
    }
    defer db.Close()

     github
    topics = []string{Selenium, Docker, Milvus}
    daysList = []int{2, 7, 45}  List of timeframes to check

    for _, days = range daysList {
        log.Println(Fetching issues for the past %d daysn, days)
        for _, topic = range topics {
            err = fetchAndStoreIssues(db, topic, days)
            if err != nil {
                log.Println(err)
                continue
            }
        }
    }

    repos = []string{prometheusprometheus, golanggo}
    for _, days = range daysList {
        log.Println(Fetching issues for the past %d daysn, days)
        for _, repo = range repos {
            issues, err = getRepoLastNDaysGitHubIssues(repo, days)
            if err != nil {
                log.Println(err)
                continue
            }

            if len(issues) == 0 {
                log.Println(No new issues found for %sn, repo)
                continue
            }

            err = insertIssues(db, issues, days)
            if err != nil {
                log.Println(Error inserting issues for %s into database %vn, repo, err)
                continue
            }

            log.Println(Successfully inserted %d issues for %s into the database.n, len(issues), repo)
        }
    }
}