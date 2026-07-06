// Copyright IBM Corp. 2024, 2025
// SPDX-License-Identifier: BUSL-1.1

package provenance

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

type gitRunner func(workingDir string, args ...string) (string, error)

func DetectBuilderID(env map[string]string) string {
	for _, key := range []string{"GITHUB_WORKFLOW_REF", "CI_JOB_URL", "CI_PIPELINE_URL", "BUILD_URL"} {
		if value := strings.TrimSpace(env[key]); value != "" {
			return value
		}
	}

	return DefaultLocalBuilderID
}

func DetectInvocationID(env map[string]string) string {
	for _, key := range []string{"GITHUB_RUN_ID", "CI_PIPELINE_ID", "CI_JOB_ID", "BUILD_ID"} {
		if value := strings.TrimSpace(env[key]); value != "" {
			return value
		}
	}

	return ""
}

func DetectGitDependency(workingDir string, env map[string]string) (ResolvedDependency, bool) {
	return detectGitDependency(workingDir, env, runGitCommand)
}

func detectGitDependency(workingDir string, env map[string]string, runner gitRunner) (ResolvedDependency, bool) {
	if dependency, ok := detectGitHubDependency(env); ok {
		return dependency, true
	}

	if dependency, ok := detectGitLabDependency(env); ok {
		return dependency, true
	}

	return detectLocalGitDependency(workingDir, runner)
}

func detectGitHubDependency(env map[string]string) (ResolvedDependency, bool) {
	repository := strings.TrimSpace(env["GITHUB_REPOSITORY"])
	commit := strings.TrimSpace(env["GITHUB_SHA"])
	if repository == "" || commit == "" {
		return ResolvedDependency{}, false
	}

	serverURL := strings.TrimRight(strings.TrimSpace(env["GITHUB_SERVER_URL"]), "/")
	if serverURL == "" {
		serverURL = "https://github.com"
	}

	uri := fmt.Sprintf("git+%s/%s", serverURL, strings.TrimLeft(repository, "/"))
	if ref := strings.TrimSpace(env["GITHUB_REF"]); ref != "" {
		uri += "@" + ref
	}

	return ResolvedDependency{
		URI: uri,
		Digest: DigestSet{
			"gitCommit": commit,
		},
	}, true
}

func detectGitLabDependency(env map[string]string) (ResolvedDependency, bool) {
	projectURL := strings.TrimSpace(env["CI_PROJECT_URL"])
	commit := strings.TrimSpace(env["CI_COMMIT_SHA"])
	if projectURL == "" || commit == "" {
		return ResolvedDependency{}, false
	}

	uri := "git+" + strings.TrimRight(projectURL, "/")
	if ref := strings.TrimSpace(env["CI_COMMIT_REF_NAME"]); ref != "" {
		uri += "@refs/heads/" + ref
	}

	return ResolvedDependency{
		URI: uri,
		Digest: DigestSet{
			"gitCommit": commit,
		},
	}, true
}

func detectLocalGitDependency(workingDir string, runner gitRunner) (ResolvedDependency, bool) {
	if runner == nil {
		return ResolvedDependency{}, false
	}

	commit, err := runner(workingDir, "rev-parse", "HEAD")
	if err != nil || commit == "" {
		return ResolvedDependency{}, false
	}

	repositoryURL, err := runner(workingDir, "config", "--get", "remote.origin.url")
	if err != nil || repositoryURL == "" {
		topLevel, topLevelErr := runner(workingDir, "rev-parse", "--show-toplevel")
		if topLevelErr != nil || topLevel == "" {
			repositoryURL = ""
		} else {
			repositoryURL = "file://" + filepath.Clean(topLevel)
		}
	}

	if repositoryURL == "" {
		return ResolvedDependency{}, false
	}

	uri := "git+" + repositoryURL
	if branch, branchErr := runner(workingDir, "rev-parse", "--abbrev-ref", "HEAD"); branchErr == nil && branch != "" && branch != "HEAD" {
		uri += "@refs/heads/" + branch
	}

	return ResolvedDependency{
		URI: uri,
		Digest: DigestSet{
			"gitCommit": commit,
		},
	}, true
}

func runGitCommand(workingDir string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	if workingDir != "" {
		command.Dir = workingDir
	}

	output, err := command.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}
