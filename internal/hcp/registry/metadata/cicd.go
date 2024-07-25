package metadata

import (
	"fmt"
	"os"
)

type GithubActions struct{}

func (g *GithubActions) Detect() error {
	_, ok := os.LookupEnv("GITHUB_ACTIONS")
	if !ok {
		return fmt.Errorf("GITHUB_ACTIONS environment variable not found")
	}
	return nil
}

func (g *GithubActions) Details() map[string]interface{} {
	env := make(map[string]interface{})
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
	return "github"
}

type GitlabCI struct{}

func (g *GitlabCI) Detect() error {
	_, ok := os.LookupEnv("GITLAB_CI")
	if !ok {
		return fmt.Errorf("GITLAB_CI environment variable not found")
	}
	return nil
}

func (g *GitlabCI) Details() map[string]interface{} {
	env := make(map[string]interface{})
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
	return "gitlab"
}

func GetCicdMetadata() map[string]interface{} {
	cicd := []MetadataProvider{
		&GithubActions{},
		&GitlabCI{},
	}

	for _, c := range cicd {
		err := c.Detect()
		if err == nil {
			return map[string]interface{}{
				"type":    c.Type(),
				"details": c.Details(),
			}
		}
	}

	return nil
}
