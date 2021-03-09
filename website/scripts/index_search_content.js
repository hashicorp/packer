require('dotenv').config()
const fs = require('fs')
const path = require('path')
const {
  indexContent,
  getDocsSearchObject,
} = require('@hashicorp/react-search/tools')
const resolveNavData = require('../components/remote-plugin-docs/utils/resolve-nav-data')
const fetchGithubFile = require('../components/remote-plugin-docs/utils/fetch-github-file')

const GITHUB_API_TOKEN = process.env.CI_GITHUB_TOKEN

// Run indexing
indexContent({ getSearchObjects })

async function getSearchObjects() {
  // Set up an array to collect all search objects
  const searchObjects = []
  // Resolve /docs, /guides, and /intro nav data, which
  // corresponds to all the content we will actually render
  // This avoids indexing non-rendered content, and partials.
  // `docs` content
  const docsNav = await resolveNavData(
    'data/docs-nav-data.json',
    'content/docs',
    {
      remotePluginsFile: 'data/docs-remote-plugins.json',
      githubToken: GITHUB_API_TOKEN,
    }
  )
  const docsObjects = await searchObjectsFromNavData(docsNav, 'docs')
  searchObjects.push(...docsObjects)
  // `guides` content
  const guidesNav = await resolveNavData(
    'data/guides-nav-data.json',
    'content/guides'
  )
  const guidesObjects = await searchObjectsFromNavData(guidesNav, 'guides')
  searchObjects.push(...guidesObjects)
  // `intro` content
  const introNav = await resolveNavData(
    'data/intro-nav-data.json',
    'content/intro'
  )
  const introObjects = await searchObjectsFromNavData(introNav, 'intro')
  searchObjects.push(...introObjects)
  // Return the collected search objects
  return searchObjects
}

/**
 * Given navData, return a flat array of search objects
 * for each content file referenced in the navData nodes
 * @param {Object[]} navData - an array of nav-data nodes, as detailed n [mktg-032](https://docs.google.com/document/d/1kYvbyd6njHFSscoE1dtDNHQ3U8IzaMdcjOS0jg87rHg)
 * @param {*} baseRoute - the base route where the navData will be rendered. For example, "docs".
 * @returns {Object[]} - an array of searchObjects to pass to Algolia. Must include an objectID property. See https://www.algolia.com/doc/api-reference/api-methods/add-objects/?client=javascript#examples.
 */
async function searchObjectsFromNavData(navData, baseRoute = '') {
  const searchObjectsFromNodes = await Promise.all(
    navData.map((n) => searchObjectsFromNavNode(n, baseRoute))
  )
  const flattenedSearchObjects = searchObjectsFromNodes.reduce(
    (acc, searchObjects) => acc.concat(searchObjects),
    []
  )
  return flattenedSearchObjects
}

/**
 * Given a single navData node, return a flat array of search objects.
 * For "leaf" nodes, this will yield an array with a single object.
 * For "branch" nodes, this may yield an array with zero or more search objects.
 * For all other nodes, this will yield an empty array.
 * @param {object} node - a nav-data nodes, as detailed in [mktg-032](https://docs.google.com/document/d/1kYvbyd6njHFSscoE1dtDNHQ3U8IzaMdcjOS0jg87rHg)
 * @param {*} baseRoute - the base route where the navData will be rendered. For example, "docs".
 * @returns {Object[]} - an array of searchObjects to pass to Algolia. Must include an objectID property. See https://www.algolia.com/doc/api-reference/api-methods/add-objects/?client=javascript#examples.
 */
async function searchObjectsFromNavNode(node, baseRoute) {
  // If this is a node, build a search object
  if (node.path) {
    //  Fetch the MDX file content
    const [err, fileString] = node.filePath
      ? //  Read local content from the filesystem
        [null, fs.readFileSync(path.join(process.cwd(), node.filePath), 'utf8')]
      : // Fetch remote content using GitHub's API
        await fetchGithubFile(node.remoteFile, GITHUB_API_TOKEN)
    if (err) throw new Error(err)
    const searchObject = await getDocsSearchObject(
      path.join(baseRoute, node.path),
      fileString
    )
    return searchObject
  }
  //  If this is a branch, recurse
  if (node.routes) return await searchObjectsFromNavData(node.routes, baseRoute)
  // Otherwise, return an empty array
  // (for direct link nodes, divider nodes)
  return []
}
