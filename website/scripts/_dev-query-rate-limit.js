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
async function getRateLimit() {
  //  Set up the GraphQL query
  const query = `query {
  rateLimit {
    limit
    cost
    remaining
    resetAt
  }
}`
  //  Set variables
  const { data } = await githubQuery({ query }, GITHUB_API_TOKEN)
  const { rateLimit } = data
  console.log({ rateLimit })
}

getRateLimit().then(() => console.log('✅ Done'))
