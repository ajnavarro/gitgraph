package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ajnavarro/gitgraph"
	"github.com/vektah/gqlgen/handler"
	git "gopkg.in/src-d/go-git.v4"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("You must specify the path to the repository and server port")
		return
	}

	repo, err := git.PlainOpen(os.Args[1])
	if err != nil {
		panic(err)
	}

	http.Handle("/", handler.Playground("Repository", "/query"))
	http.Handle("/query", handler.GraphQL(gitgraph.MakeExecutableSchema(gitgraph.NewRepoResolvers(repo))))
	log.Fatal(http.ListenAndServe(":"+os.Args[2], nil))
}
