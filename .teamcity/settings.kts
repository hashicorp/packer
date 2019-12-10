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
            type = "OAuthProvider"
            param("providerType", "GitHub")
            param("displayName", "GitHub.com")
            param("gitHubUrl", "https://github.com/")
            param("clientId", "1abfd46417d7795298a1")
            param("secure:clientSecret", "credentialsJSON:5fe99dc3-4d1d-4fd6-9f5c-e87fbcbd9a4e")
            param("defaultTokenScope", "public_repo,repo,repo:status,write:repo_hook")
        }
        feature {
            type = "IssueTracker"
            param("name", "packer-builder-vsphere")
            param("type", "GithubIssues")
            param("repository", "https://github.com/jetbrains-infra/packer-builder-vsphere")
            param("authType", "anonymous")
            param("pattern", """#(\d+)""")
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
        password("env.VSPHERE_PASSWORD", "credentialsJSON:d5e7ac7f-357b-464a-b2fa-ddd4c433b22b")
    }

    steps {
        script {
            name = "Build"
            scriptContent = "make build -j 3"
            dockerImage = golangImage
            dockerPull = true
        }

        dockerCompose {
            name = "Start VPN tunnel"
            file = "teamcity-services.yml"
        }

        script {
            name = "Test"
            scriptContent = "make test | go-test-teamcity"
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
                    token = "credentialsJSON:5ead3bb1-c370-4589-beb8-24f8d02c36bc"
                }
            }
        }
        pullRequests {
            provider = github {
                authType = token {
                    token = "credentialsJSON:5ead3bb1-c370-4589-beb8-24f8d02c36bc"
                }
                filterAuthorRole = PullRequests.GitHubRoleFilter.EVERYBODY
            }
        }
    }

    triggers {
        vcs {
            triggerRules = """
                -:*.md
                -:.teamcity/
            """.trimIndent()
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
