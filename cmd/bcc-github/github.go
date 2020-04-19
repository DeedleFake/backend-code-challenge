package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	_ "github.com/lib/pq"

	"github.com/DeedleFake/backend-code-challenge/bcc"
	"github.com/jmoiron/sqlx"
)

var httpClient = &http.Client{
	Timeout: 60 * time.Second,
}

type GitHubEvent struct {
	Type      string    `json:"type"`
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	Payload   struct {
		Action     string `json:"action"`
		Repository struct {
			FullName string `json:"full_name"`
		} `json:"repository"`
		PullRequest struct {
			Merged bool `json:"merged"`
		} `json:"pull_request"`
		Number int    `json:"number"`
		Size   int    `json:"size"`
		Head   string `json:"head"`
	} `json:"payload"`
}

func getEvents(user string, token string) ([]GitHubEvent, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/users/%v/events", user), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %v", token))
	}

	rsp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("get events: %w", err)
	}
	defer rsp.Body.Close()

	var events []GitHubEvent
	err = json.NewDecoder(rsp.Body).Decode(&events)
	if err != nil {
		return nil, fmt.Errorf("decode: %w", err)
	}
	return events, nil
}

func addEvents(db *sqlx.DB, userID int, ghuser string, token string) error {
	events, err := getEvents(ghuser, token)
	if err != nil {
		log.Fatalf("Error getting events: %v", err)
	}

	for _, event := range events {
		gh := bcc.GitHubEvent{
			ID:        event.ID,
			CreatedAt: event.CreatedAt,
			UserID:    userID,
		}

		switch event.Type {
		case "RepositoryEvent":
			if event.Payload.Action != "created" {
				continue
			}

			gh.Type = "create_repo"
			gh.RepoName = event.Payload.Repository.FullName

		case "PullRequestEvent":
			switch event.Payload.Action {
			case "opened":
				gh.Type = "open_pr"

			case "closed":
				if !event.Payload.PullRequest.Merged {
					continue
				}
				gh.Type = "merge_pr"

			default:
				continue
			}

			gh.RepoName = event.Payload.Repository.FullName
			gh.PRNumber = &event.Payload.Number

		case "PushEvent":
			gh.Type = "push_commits"
			gh.RepoName = event.Payload.Repository.FullName
			gh.NumCommits = &event.Payload.Size
			gh.Head = &event.Payload.Head
		}

		err := bcc.AddGitHubEvent(db, gh)
		if err != nil {
			log.Printf("Error adding %v: %v", gh.ID, err)
		}
	}

	return nil
}

func main() {
	dbaddr := flag.String("dbaddr", "localhost", "Database address")
	dbuser := flag.String("dbuser", "postgres", "Database user")
	dbpass := flag.String("dbpass", "", "Database password")
	dbname := flag.String("dbname", "bcc", "Database name")
	token := flag.String("token", "", "GitHub OAuth 2 token to increase rate limit")
	flag.Parse()

	db, err := sqlx.Open("postgres", fmt.Sprintf(
		"postgres://%v:%v@%v/%v?sslmode=disable",
		*dbuser,
		*dbpass,
		*dbaddr,
		*dbname,
	))
	if err != nil {
		log.Fatalf("Failed to open database connection: %v", err)
	}
	defer db.Close()

	rows, err := db.Queryx(`SELECT id, github_username FROM users WHERE github_username IS NOT NULL`)
	if err != nil {
		log.Fatalf("Failed to get user list: %v", err)
	}
	defer rows.Close()

	var failed bool
	var wg sync.WaitGroup
	for rows.Next() {
		var user struct {
			ID         int    `db:"id"`
			GHUsername string `db:"github_username"`
		}
		err := rows.StructScan(&user)
		if err != nil {
			log.Fatalf("Failed to scan: %v", err)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()

			err := addEvents(db, user.ID, user.GHUsername, *token)
			if err != nil {
				log.Printf("Failed to add events for %q (%v): %v", user.GHUsername, user.ID, err)
				failed = true
			}
		}()
	}
	wg.Wait()

	if failed {
		os.Exit(1)
	}
}