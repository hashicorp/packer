const fetch = require('isomorphic-unfetch')

const GITHUB_API_TOKEN = process.env.GITHUB_API_TOKEN

async function githubQuery(body, token) {
  const result = await fetch('https://api.github.com/graphql', {
    method: 'POST',
    headers: {
      Authorization: `bearer ${token}`,
      ContentType: 'application/json',
    },
    body: JSON.stringify(body),
  })
  return await result.json()
}

//  Fetch a file from GitHub using the GraphQL API
async function getGithubFile({ repo, branch, filePath }) {
  const [repo_owner, repo_name] = repo.split('/')
  //  Set up the GraphQL query
  // (usually we can keep this in a separate file, and rely on a
  // plaintext loader we've set up in our NextJS config, but we need
  // to fetch remote content when indexing it, which happens outside
  // NextJS, so unfortunately it seems this has to be inlined)
  const query = `
query($repo_name: String!, $repo_owner: String!, $object_expression: String!) {
  repository(name: $repo_name, owner: $repo_owner) {
    object(expression: $object_expression) {
      ... on Blob {
        text
      }
    }
  }
}
`
  //  Set variables
  const variables = {
    repo_name,
    repo_owner,
    object_expression: `${branch}:${filePath}`,
  }
  // Query the GitHub API, and parse the navigation data
  const result = await githubQuery({ query, variables }, GITHUB_API_TOKEN)
  try {
    const fileText = result.data.repository.object.text
    return [null, fileText]
  } catch (e) {
    const errorMsg = `Could not fetch remote file text from "${
      variables.object_expression
    }" in "${repo_owner}/${repo_name}". Received instead:\n\n${JSON.stringify(
      result,
      null,
      2
    )}`
    return [errorMsg, null]
  }
}

function memoize(method) {
  let cache = {}

  return async function () {
    let args = JSON.stringify(arguments[0])
    if (!cache[args]) {
      cache[args] = method.apply(this, arguments)
    }
    return cache[args]
  }
}

module.exports = memoize(getGithubFile)
