// bcc-github updates the database with public GitHub events.
//
// This command is designed to be run as a cron job. It searches the
// database for users that have linked their GitHub profiles and gets
// event data from the GitHub API, adding it to the database so that
// it can show up in those users' timelines. It uses the timestamps
// that GitHub provides, so events will be sorted correctly in the
// timeline relative to other entries.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	_ "github.com/lib/pq"

	"github.com/DeedleFake/backend-code-challenge/bcc"
	"github.com/jmoiron/sqlx"
)

var httpClient = &http.Client{
	Timeout: 60 * time.Second,
}

// GitHubEvent is a minimal structure for decoding the returned data
// from the GitHub API.
type GitHubEvent struct {
	Type      string    `json:"type"`
	ID        uint64    `json:"id,string"`
	CreatedAt time.Time `json:"created_at"`
	Repo      struct {
		Name string `json:"name"`
	} `json:"repo"`
	Payload struct {
		Action      string `json:"action"`
		PullRequest struct {
			Merged bool `json:"merged"`
		} `json:"pull_request"`
		Number uint64 `json:"number"`
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

func addEvents(db *sqlx.DB, userID uint64, ghuser string, token string) error {
	events, err := getEvents(ghuser, token)
	if err != nil {
		return fmt.Errorf("get events: %w", err)
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
			gh.RepoName = event.Repo.Name

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

			gh.RepoName = event.Repo.Name
			gh.PRNumber = &event.Payload.Number

		case "PushEvent":
			gh.Type = "push_commits"
			gh.RepoName = event.Repo.Name
			gh.NumCommits = &event.Payload.Size
			gh.Head = &event.Payload.Head

		default:
			continue
		}

		err := bcc.AddGitHubEvent(db, gh)
		if err != nil {
			return fmt.Errorf("add %v: %w", gh.ID, err)
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

	var failed uint32
	var wg sync.WaitGroup
	for rows.Next() {
		var user struct {
			ID         uint64 `db:"id"`
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
				atomic.StoreUint32(&failed, 1)
			}
		}()
	}
	wg.Wait()

	if failed != 0 {
		os.Exit(1)
	}
}
