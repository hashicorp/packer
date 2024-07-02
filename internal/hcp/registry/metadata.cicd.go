package registry

import (
	"fmt"
	"os"
)

type CICD interface {
	Detect() bool
	Env() map[string]string
	Type() string
}

type GithubActions struct{}

func (g *GithubActions) Detect() bool {
	_, ok := os.LookupEnv("GITHUB_ACTIONS")
	return ok
}

func (g *GithubActions) Env() map[string]string {
	env := make(map[string]string)
	keys := []string{
		"GITHUB_REPOSITORY",
		"GITHUB_REPOSITORY_ID",
		"GITHUB_WORKFLOW_URL",
		"GITHUB_SHA",
		"GITHUB_REF",
		"GITHUB_ACTOR",
		"GITHUB_ACTOR_ID",
		"GITHUB_TRIGGERING_ACTOR",
		"GITHUB_EVENT_NAME",
		"GITHUB_JOB",
	}

	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			env[key] = value
		}
	}

	env["GITHUB_WORKFLOW_URL"] = fmt.Sprintf("%s/%s/actions/runs/%s", os.Getenv("GITHUB_SERVER_URL"), os.Getenv("GITHUB_REPOSITORY"), os.Getenv("GITHUB_RUN_ID"))
	return env
}

func (g *GithubActions) Type() string {
	return "github-actions"
}

type GitlabCI struct{}

func (g *GitlabCI) Detect() bool {
	_, ok := os.LookupEnv("GITLAB_CI")
	return ok
}

func (g *GitlabCI) Env() map[string]string {
	env := make(map[string]string)
	keys := []string{
		"CI_PROJECT_NAME",
		"CI_PROJECT_ID",
		"CI_PROJECT_URL",
		"CI_COMMIT_SHA",
		"CI_COMMIT_REF_NAME",
		"GITLAB_USER_NAME",
		"GITLAB_USER_ID",
		"CI_PIPELINE_SOURCE",
		"CI_PIPELINE_URL",
		"CI_JOB_URL",
		"CI_SERVER_NAME",
		"CI_REGISTRY_IMAGE",
	}

	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			env[key] = value
		}
	}

	return env
}

func (g *GitlabCI) Type() string {
	return "gitlab-ci"
}

func GetCicdMetadata() map[string]interface{} {
	cicd := []CICD{
		&GithubActions{},
		&GitlabCI{},
	}

	for _, c := range cicd {
		if c.Detect() {
			return map[string]interface{}{
				"type":    c.Type(),
				"details": c.Env(),
			}
		}
	}

	return nil
}
