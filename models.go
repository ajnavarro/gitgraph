package gitgraph

type Reference struct {
	Hash string
	Name string
}

type Commit struct {
	Hash        string
	Message     string
	AuthorName  string
	AuthorEmail string
}
