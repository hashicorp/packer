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

type BitbucketPipelines struct{}

func (b *BitbucketPipelines) Detect() error {
	_, ok := os.LookupEnv("BITBUCKET_BUILD_NUMBER")
	if !ok {
		return fmt.Errorf("BITBUCKET_BUILD_NUMBER environment variable not found")
	}
	return nil
}

func (b *BitbucketPipelines) Details() map[string]interface{} {
	env := make(map[string]interface{})
	keys := []string{
		"BITBUCKET_REPO_FULL_NAME",
		"BITBUCKET_REPO_UUID",
		"BITBUCKET_WORKSPACE",
		"BITBUCKET_COMMIT",
		"BITBUCKET_BRANCH",
		"BITBUCKET_TAG",
		"BITBUCKET_BUILD_NUMBER",
		"BITBUCKET_PIPELINE_UUID",
		"BITBUCKET_STEP_UUID",
		"BITBUCKET_DEPLOYMENT_ENVIRONMENT",
		"BITBUCKET_PR_ID",
		"BITBUCKET_PR_DESTINATION_BRANCH",
		"BITBUCKET_PROJECT_KEY",
		"BITBUCKET_PROJECT_UUID",
	}

	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			env[key] = value
		}
	}

	return env
}

func (b *BitbucketPipelines) Type() string {
	return "bitbucket"
}

type JenkinsCI struct{}

func (g *JenkinsCI) Detect() error {
	_, ok := os.LookupEnv("JENKINS_URL")
	if !ok {
		return fmt.Errorf("JENKINS_URL environment variable not found")
	}
	return nil
}

func (g *JenkinsCI) Details() map[string]interface{} {
	env := make(map[string]interface{})
	keys := []string{
		"JENKINS_URL",
		"BUILD_URL",
		"NODE_NAME",
		"JOB_NAME",
		"JOB_URL",
		"BUILD_NUMBER",
		"BUILD_ID",
		"BUILD_TAG",
		"WORKSPACE",
		"BUILD_CAUSE",
		"GIT_COMMIT",
		"GIT_BRANCH",
		"GIT_URL",
		"GIT_AUTHOR_NAME",
		"GIT_COMMITTER_EMAIL",
		"GIT_PREVIOUS_SUCCESSFUL_COMMIT",
	}

	for _, key := range keys {
		if value, ok := os.LookupEnv(key); ok {
			env[key] = value
		}
	}

	return env
}

func (g *JenkinsCI) Type() string {
	return "jenkins"
}

func GetCicdMetadata() map[string]interface{} {
	cicd := []MetadataProvider{
		&JenkinsCI{},
		&GithubActions{},
		&GitlabCI{},
		&BitbucketPipelines{},
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
