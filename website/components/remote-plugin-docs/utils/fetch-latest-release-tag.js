async function fetchLatestReleaseTag(repo) {
  const latestReleaseUrl = `https://github.com/${repo}/releases/latest`
  const redirectedUrl = await getRedirectedUrl(latestReleaseUrl)
  if (!redirectedUrl) return false
  const latestTag = redirectedUrl.match(/tag\/(.*)/)[1]
  return latestTag
}

async function getRedirectedUrl(url) {
  return new Promise((resolve, reject) => {
    const https = require('https')
    const req = https.request(url, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400) {
        resolve(res.headers.location)
      } else {
        resolve(false)
      }
    })
    req.on('error', reject)
    req.end()
  })
}

module.exports = fetchLatestReleaseTag
