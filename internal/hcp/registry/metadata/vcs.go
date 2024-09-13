package metadata

import (
	"log"
	"os"

	gt "github.com/go-git/go-git/v5"
)

type MetadataProvider interface {
	Detect() error
	Details() map[string]interface{}
	Type() string
}

type Git struct {
	repo *gt.Repository
}

func (g *Git) Detect() error {
	wd, err := os.Getwd()
	if err != nil {
		log.Printf("[ERROR] unable to retrieve current directory: %s", err)
		return err
	}

	repo, err := gt.PlainOpenWithOptions(wd, &gt.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return err
	}

	g.repo = repo
	return nil
}

func (g *Git) hasUncommittedChanges() bool {
	worktree, err := g.repo.Worktree()
	if err != nil {
		log.Printf("[ERROR] failed to get the git worktree: %s", err)
		return false
	}

	status, err := worktree.Status()
	if err != nil {
		log.Printf("[ERROR] failed to get the git worktree status: %s", err)
		return false
	}
	return !status.IsClean()
}

func (g *Git) Type() string {
	return "git"
}

func (g *Git) Details() map[string]interface{} {
	headRef, err := g.repo.Head()
	if err != nil {
		log.Printf("[ERROR] failed to get reference to git HEAD: %s", err)
		return nil
	}

	resp := map[string]interface{}{
		"ref": headRef.Name().Short(),
	}

	commit, err := g.repo.CommitObject(headRef.Hash())
	if err != nil {
		log.Printf("[ERROR] failed to get the git commit hash: %s", err)
	} else {
		resp["commit"] = commit.Hash.String()
		resp["author"] = commit.Author.Name + " <" + commit.Author.Email + ">"
	}

	resp["has_uncommitted_changes"] = g.hasUncommittedChanges()
	return resp
}

func GetVcsMetadata() map[string]interface{} {
	vcsSystems := []MetadataProvider{
		&Git{},
	}

	for _, vcs := range vcsSystems {
		err := vcs.Detect()
		if err == nil {
			return map[string]interface{}{
				"type":    vcs.Type(),
				"details": vcs.Details(),
			}
		}
	}

	return nil
}
