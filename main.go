package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/shurcooL/githubql"
	"golang.org/x/oauth2"
)

type (
	configuration struct {
		GITHUB_KEY string
	}
	edge struct {
		Cursor githubql.String
		Node   struct {
			Repository struct {
				CreatedAt   githubql.DateTime
				Description githubql.String
				DiskUsage   githubql.Int
				Name        githubql.String
				UpdatedAt   githubql.DateTime
				Url         githubql.String
			} `graphql:"... on Repository"`
		}
	}
)

var (
	numberPages        int
	numberRepositories int
	pageCursor         string
	pageSize           = 100
	queryInitial       struct {
		Search struct {
			RepositoryCount githubql.Int
			Edges           []edge
		} `graphql:"search(query: $queryString, type: REPOSITORY, first:$pageSize)"`
	}
	queryInitialVariables = map[string]interface{}{
		"queryString": githubql.String(queryString),
		"pageSize":    githubql.Int(pageSize),
	}
	queryPaginate struct {
		Search struct {
			Edges []edge
		} `graphql:"search(query: $queryString, type: REPOSITORY, first:$pageSize, after:$pageCursor)"`
	}
	queryString = "language:PowerShell"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func writeResults(file *os.File, edges []edge) {
	for _, edge := range edges {
		// ADVANCE CURSOR
		pageCursor = string(edge.Cursor)
		// WRITE RESULTS
		file.WriteString(string(edge.Node.Repository.Url))
		file.WriteString("\n")
		file.Sync()
	}
}

func main() {
	// OPEN JSON CONFIG FILE AND DECODE
	configFile, err := os.Open("config.json")
	check(err)
	fmt.Println("Config file opened.")

	decoder := json.NewDecoder(configFile)
	config := configuration{}
	err = decoder.Decode(&config)
	check(err)
	fmt.Println("Config file decoded.")

	// CONNECT TO GITHUB GRAPHQL API
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: config.GITHUB_KEY},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubql.NewClient(httpClient)

	// MAKE OUTPUT DIRECTORY IF IT DOES NOT EXIST
	if _, err := os.Stat("output"); err != nil {
		err = os.Mkdir("output", 0777)
		check(err)
		fmt.Println("Output directory created.")
	} else {
		check(err)
	}

	// OPEN FILE TO BEGIN WRITING
	file, err := os.OpenFile("output/url", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0777)
	check(err)
	fmt.Println("Local results file ready.")
	// CLOSE FILE WHEN DONE
	defer file.Close()

	// SUBMIT INITIAL QUERY FOR FIRST PAGE OF RESULTS
	err = client.Query(context.Background(), &queryInitial, queryInitialVariables)
	check(err)
	fmt.Println("Query #1 successful.")

	// RETRIEVE TOTAL NUMBER OF REPOSITORIES
	numberRepositories = int(queryInitial.Search.RepositoryCount)
	// CALCULATE NUMBER OF PAGES TO PAGINATE THROUGH
	numberPages = numberRepositories / pageSize

	writeResults(file, queryInitial.Search.Edges)

	// WRITE ADDITIONAL PAGES OF RESULTS
	for i := 0; i < numberPages; i++ {
		time.Sleep(1000 * time.Millisecond)

		// INSTANTIATE NEW VARIABLES WITH UPDATED PAGECURSOR
		queryPaginateVariables := map[string]interface{}{
			"queryString": githubql.String(queryString),
			"pageCursor":  githubql.String(pageCursor),
			"pageSize":    githubql.Int(pageSize),
		}

		// SUBMIT NEW QUERY FOR FOLLOW ON PAGE
		err := client.Query(context.Background(), &queryPaginate, queryPaginateVariables)
		if err != nil {
			src = oauth2.StaticTokenSource(
				&oauth2.Token{AccessToken: config.GITHUB_KEY},
			)
			httpClient = oauth2.NewClient(context.Background(), src)
			client = githubql.NewClient(httpClient)
		}
		// check(err)
		fmt.Printf("Query #%v successful.", i+2)
		fmt.Printf("\n")

		writeResults(file, queryPaginate.Search.Edges)
	}
}
