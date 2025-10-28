package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/machinebox/graphql"
)

// GraphQL APIからのレスポンスをマッピングするための構造体
type ResponseData struct {
	Repository struct {
		PullRequests struct {
			Nodes []PullRequest
		}
	}
}

type PullRequest struct {
	Number    int
	Title     string
	Author    struct {
		Login string
	}
	CreatedAt time.Time
	MergedAt  *time.Time
	Commits   struct {
		Nodes []struct {
			Commit struct {
				CommittedDate time.Time
			}
		}
	} `graphql:"commits(first: 1)"`
	Reviews struct {
		Nodes []struct {
			Author struct {
				Login string
			}
			State       string
			SubmittedAt time.Time
		}
	} `graphql:"reviews(first: 10)"`
	Comments struct {
		Nodes []struct {
			Author struct {
				Login string
			}
			Body      string
			CreatedAt time.Time
		}
	} `graphql:"comments(first: 20)"`
}

func main() {
	// .envファイルから環境変数を読み込む
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	githubToken := os.Getenv("GITHUB_TOKEN")
	if githubToken == "" {
		log.Fatal("Error: GITHUB_TOKEN is not set.")
	}

	// コマンドライン引数をパースする
	prCount := flag.Int("n", 10, "Number of pull requests to fetch")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Usage: go run main.go [-n <number>] <owner>/<repository>")
	}
	repoParts := strings.Split(args[0], "/")
	if len(repoParts) != 2 {
		log.Fatal("Invalid repository format. Use <owner>/<repository>.")
	}
	owner, repo := repoParts[0], repoParts[1]

	// GraphQLクライアントを作成
	client := graphql.NewClient("https://api.github.com/graphql")

	// GraphQLクエリを定義
	query := `
        query($owner: String!, $repo: String!, $prCount: Int!) {
          repository(owner: $owner, name: $repo) {
            pullRequests(last: $prCount, states: [MERGED, CLOSED]) {
              nodes {
                number
                title
                author {
                  login
                }
                createdAt
                mergedAt
                commits(first: 1) {
                  nodes {
                    commit {
                      committedDate
                    }
                  }
                }
                reviews(first: 10, states: [COMMENTED, APPROVED, CHANGES_REQUESTED]) {
                  nodes {
                    author {
                      login
                    }
                    state
                    submittedAt
                  }
                }
                comments(first: 20) {
                    nodes {
                        author {
                            login
                        }
                        body
                        createdAt
                    }
                }
              }
            }
          }
        }
    `

	// リクエストを作成
	req := graphql.NewRequest(query)
	req.Var("owner", owner)
	req.Var("repo", repo)
	req.Var("prCount", *prCount)
	req.Header.Set("Authorization", "Bearer "+githubToken)

	// クエリを実行
	var respData ResponseData
	ctx := context.Background()
	if err := client.Run(ctx, req, &respData); err != nil {
		log.Fatalf("Error running GraphQL query: %v", err)
	}

	// Markdownレポートを生成
	fmt.Printf("# Pull Request Analysis Report for %s/%s\n\n", owner, repo)
	for _, pr := range respData.Repository.PullRequests.Nodes {
		printPullRequestMarkdown(pr)
	}
}

func printPullRequestMarkdown(pr PullRequest) {
	fmt.Printf("## PR #%d: %s\n", pr.Number, pr.Title)
	fmt.Printf("- **Author:** %s\n", pr.Author.Login)
	fmt.Printf("- **Created at:** %s\n", pr.CreatedAt.Format(time.RFC3339))
	if pr.MergedAt != nil {
		fmt.Printf("- **Merged at:** %s\n", pr.MergedAt.Format(time.RFC3339))
		fmt.Printf("- **Lead Time:** %s\n", pr.MergedAt.Sub(pr.CreatedAt).Round(time.Minute))
	} else {
		fmt.Println("- **Status:** Closed without merging")
	}

	// サイクルタイムの計算と表示
	if len(pr.Commits.Nodes) > 0 {
		firstCommitDate := pr.Commits.Nodes[0].Commit.CommittedDate
		if pr.MergedAt != nil {
			fmt.Printf("- **Commit to Merge Time:** %s\n", pr.MergedAt.Sub(firstCommitDate).Round(time.Minute))
		}
	}
	if len(pr.Reviews.Nodes) > 0 {
		firstReviewDate := pr.Reviews.Nodes[0].SubmittedAt
		fmt.Printf("- **Time to First Review:** %s\n", firstReviewDate.Sub(pr.CreatedAt).Round(time.Minute))
	}
	fmt.Println()

	fmt.Println("### Reviews")
	if len(pr.Reviews.Nodes) == 0 {
		fmt.Println("- No reviews")
	} else {
		for _, review := range pr.Reviews.Nodes {
			fmt.Printf("- **%s:** %s - %s\n", review.Author.Login, review.State, review.SubmittedAt.Format(time.RFC3339))
		}
	}
	fmt.Println()

	fmt.Println("### Comments")
	if len(pr.Comments.Nodes) == 0 {
		fmt.Println("- No comments")
	} else {
        // 簡単にするため、最初の5件のみ表示
        limit := 5
        if len(pr.Comments.Nodes) < limit {
            limit = len(pr.Comments.Nodes)
        }
		for _, comment := range pr.Comments.Nodes[:limit] {
            // コメント本文は改行を潰して短く表示
            body := strings.ReplaceAll(comment.Body, "\n", " ")
            if len(body) > 70 {
                body = body[:67] + "..."
            }
			fmt.Printf("- **%s:** \"%s\" - %s\n", comment.Author.Login, body, comment.CreatedAt.Format(time.RFC3339))
		}
	}

	fmt.Println("\n---")
}
