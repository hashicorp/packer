const fetch = require('node-fetch')

/**
 * @param {string} repo
 * @returns {Promise<string|boolean>}
 */
async function fetchLatestReleaseTag(repo) {
  const latestReleaseUrl = `https://github.com/${repo}/releases/latest`
  let res = await fetch(latestReleaseUrl)

  if (res.status !== 200) {
    console.error(
      `failed to fetch: ${latestReleaseUrl}`,
      res.status,
      res.statusText,
      res.body
    )
    return false
  }

  const matches = res.url.match(/tag\/(.*)/)

  if (!matches) {
    console.error(`failed to parse tag from: ${res.url}`)
    return false
  }

  return matches[1]
}

module.exports = fetchLatestReleaseTag
