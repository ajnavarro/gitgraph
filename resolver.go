//go:generate gorunpkg github.com/vektah/gqlgen -out generated.go

package gitgraph

import (
	context "context"
	"io"

	"github.com/araddon/qlbridge/datasource"
	"github.com/araddon/qlbridge/expr"
	"github.com/araddon/qlbridge/value"
	"github.com/araddon/qlbridge/vm"
	git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

type RepoResolvers struct {
	repo *git.Repository
}

func NewRepoResolvers(repo *git.Repository) *RepoResolvers {
	return &RepoResolvers{repo}
}

// Query_references returns all the references of a repository
func (r *RepoResolvers) Query_references(ctx context.Context, filter string) ([]Reference, error) {
	node, err := expr.ParseExpression(filter)
	if err != nil {
		return nil, err
	}

	rIter, err := r.repo.References()
	if err != nil {
		return nil, err
	}
	var result []Reference
	for {
		ref, err := rIter.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		ctx := datasource.NewContextSimpleNative(map[string]interface{}{
			"name": ref.Name().String(),
			"hash": ref.Hash().String(),
		})

		// TODO better handling
		toAdd, _ := vm.Eval(ctx, node)
		if toAdd != nil && toAdd.Type() == value.BoolType && toAdd.Value() == true {
			result = append(result, Reference{
				Hash: ref.Hash().String(),
				Name: ref.Name().String(),
			})
		}
	}

	return result, nil
}

// Query_commits returns all the commits of a repository
func (r *RepoResolvers) Query_commits(ctx context.Context, filter string) ([]Commit, error) {
	cIter, err := r.repo.CommitObjects()
	if err != nil {
		return nil, err
	}

	return r.getCommitsResult(cIter, filter)
}

func (r *RepoResolvers) Query_reference(ctx context.Context, name string) (*Reference, error) {
	ref, err := r.repo.Reference(plumbing.ReferenceName(name), true)
	if err != nil {
		return nil, err
	}

	return &Reference{
		Hash: ref.Hash().String(),
		Name: ref.Name().String(),
	}, nil
}

// Reference_commits returns all the commits in a specific reference
func (r *RepoResolvers) Reference_commits(ctx context.Context, it *Reference, filter string) ([]Commit, error) {
	ref, err := r.repo.Reference(plumbing.ReferenceName(it.Name), true)
	if err != nil {
		return nil, err
	}

	cIter, err := r.repo.Log(&git.LogOptions{
		From: ref.Hash(),
	})

	if err != nil {
		return nil, err
	}

	return r.getCommitsResult(cIter, filter)
}

func (r *RepoResolvers) getCommitsResult(cIter object.CommitIter, filter string) ([]Commit, error) {
	node, err := expr.ParseExpression(filter)
	if err != nil {
		return nil, err
	}

	var result []Commit
	for {
		commit, err := cIter.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}
		ctx := datasource.NewContextSimpleNative(map[string]interface{}{
			"hash":        commit.Hash.String(),
			"authorName":  commit.Author.Name,
			"authorEmail": commit.Author.Email,
			"message":     commit.Message,
		})

		// TODO better handling
		toAdd, _ := vm.Eval(ctx, node)
		if toAdd != nil && toAdd.Type() == value.BoolType && toAdd.Value() == true {
			result = append(result, Commit{
				Hash:        commit.Hash.String(),
				Message:     commit.Message,
				AuthorName:  commit.Author.Name,
				AuthorEmail: commit.Author.Email,
			})
		}
	}

	return result, nil
}
