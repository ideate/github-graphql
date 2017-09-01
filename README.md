# github-graphql
GitHub GraphQL API Server

## Update Go in C9
https://community.c9.io/t/writing-a-go-app/1725
sudo rm -rf /opt/go
wget https://storage.googleapis.com/golang/go1.9.linux-amd64.tar.gz
sudo tar -C /opt -xvf go1.9.linux-amd64.tar.gz
go version

## Install Packages
go get -u golang.org/x/oauth2
go get -u github.com/shurcooL/githubql

## Run App
go run main.go 

## Sample Query
```
query {
	search (query: "language:PowerShell", type: REPOSITORY, first:100){
		repositoryCount
    edges {
      cursor
      node {
				... on Repository {
          id
          name
          description
          url
          owner {
            login
          }
          createdAt
          updatedAt
          diskUsage
        }       
      }
    }
  }
}
```