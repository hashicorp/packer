import jetbrains.buildServer.configs.kotlin.v2018_2.*
import jetbrains.buildServer.configs.kotlin.v2018_2.buildFeatures.PullRequests
import jetbrains.buildServer.configs.kotlin.v2018_2.buildFeatures.commitStatusPublisher
import jetbrains.buildServer.configs.kotlin.v2018_2.buildFeatures.pullRequests
import jetbrains.buildServer.configs.kotlin.v2018_2.buildSteps.dockerCompose
import jetbrains.buildServer.configs.kotlin.v2018_2.buildSteps.script
import jetbrains.buildServer.configs.kotlin.v2018_2.triggers.vcs
import jetbrains.buildServer.configs.kotlin.v2018_2.vcs.GitVcsRoot

version = "2018.2"

project {
    description = "https://github.com/jetbrains-infra/packer-builder-vsphere"

    vcsRoot(GitHub)
    buildType(Build)

    features {
        feature {
            type = "IssueTracker"
            param("type", "GithubIssues")
            param("repository", "https://github.com/jetbrains-infra/packer-builder-vsphere")
        }
    }
}

object GitHub : GitVcsRoot({
    name = "packer-builder-vsphere"
    url = "https://github.com/jetbrains-infra/packer-builder-vsphere"
    branch = "master"
    branchSpec = "+:refs/heads/(*)"
    userNameStyle = GitVcsRoot.UserNameStyle.FULL
})

object Build : BuildType({
    val golangImage = "jetbrainsinfra/golang:1.11.4"

    name = "Build"

    vcs {
        root(GitHub)
    }

    requirements {
        equals("docker.server.osType", "linux")
        exists("dockerCompose.version")

        doesNotContain("teamcity.agent.name", "ubuntu-single-build")
    }

    params {
        param("env.GOPATH", "%teamcity.build.checkoutDir%/build/modules")
        param("env.GOCACHE", "%teamcity.build.checkoutDir%/build/cache")

        password("env.VPN_PASSWORD", "credentialsJSON:8c355e81-9a26-4788-8fea-c854cd646c35")
        param   ("env.VSPHERE_USERNAME", """vsphere65.test\teamcity""")
        password("env.VSPHERE_PASSWORD", "credentialsJSON:3e99d6c8-b66f-410a-a865-eaf1b12664ad")
    }

    steps {
        script {
            name = "Build"
            scriptContent = "./build.sh"
            dockerImage = golangImage
            dockerPull = true
        }

        dockerCompose {
            name = "Start VPN tunnel"
            file = "teamcity-services.yml"
        }

        script {
            name = "Test"
            scriptContent = """
                set -eux
                
                go test -c ./driver
                go test -c ./iso
                go test -c ./clone
                
                ./test.sh | go-test-teamcity
            """.trimIndent()
            dockerImage = golangImage
            dockerPull = true
            dockerRunParameters = "--network=container:vpn"
        }
        script {
            name = "gofmt"
            executionMode = BuildStep.ExecutionMode.RUN_ON_FAILURE
            scriptContent = "./gofmt.sh"
            dockerImage = golangImage
            dockerPull = true
        }
    }

    features {
        commitStatusPublisher {
            publisher = github {
                githubUrl = "https://api.github.com"
                authType = personalToken {
                    token = "credentialsJSON:95bbfc46-3141-4bed-86ec-f8ec751f3e94"
                }
            }
            param("github_oauth_user", "mkuzmin")
        }
        pullRequests {
            provider = github {
                authType = token {
                    token = "credentialsJSON:39727f26-62ed-4152-ab9a-f6845076a979"
                }
                filterAuthorRole = PullRequests.GitHubRoleFilter.EVERYBODY
            }
        }
    }

    triggers {
        vcs {
            triggerRules = "-:*.md"
            branchFilter = """
                +:*
                -:temp-*
                -:pull/*
            """.trimIndent()
        }
    }
    maxRunningBuilds = 2

    artifactRules = "bin/* => packer-builder-vsphere-%build.number%.zip"
    allowExternalStatus = true
})
