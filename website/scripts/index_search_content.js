require('dotenv').config()
const fs = require('fs')
const path = require('path')
const {
  indexContent,
  getDocsSearchObject,
} = require('@hashicorp/react-search/tools')
const resolveNavData = require('../components/remote-plugin-docs/utils/resolve-nav-data')

// Run indexing
indexContent({ getSearchObjects })

async function getSearchObjects() {
  // Resolve /docs, /guides, and /intro nav data, which
  // corresponds to all the content we will actually render
  // This avoids indexing non-rendered content, and partials.
  // Fetch objects for `docs` content
  async function fetchDocsObjects() {
    const navFile = 'data/docs-nav-data.json'
    const contentDir = 'content/docs'
    const opts = { remotePluginsFile: 'data/docs-remote-plugins.json' }
    const navData = await resolveNavData(navFile, contentDir, opts)
    return await searchObjectsFromNavData(navData, 'docs')
  }
  // Fetch objects for `guides` content
  async function fetchGuidesObjects() {
    const navFile = 'data/guides-nav-data.json'
    const contentDir = 'content/guides'
    const navData = await resolveNavData(navFile, contentDir)
    return await searchObjectsFromNavData(navData, 'guides')
  }
  // Fetch objects for `intro` content
  async function fetchIntroObjects() {
    const navFile = 'data/intro-nav-data.json'
    const contentDir = 'content/intro'
    const navData = await resolveNavData(navFile, contentDir)
    return await searchObjectsFromNavData(navData, 'intro')
  }
  // Collect, flatten and return the collected search objects
  const searchObjects = (
    await Promise.all([
      fetchDocsObjects(),
      fetchGuidesObjects(),
      fetchIntroObjects(),
    ])
  ).reduce((acc, array) => acc.concat(array), [])
  return searchObjects
}

/**
 * Given navData, return a flat array of search objects
 * for each content file referenced in the navData nodes
 * @param {Object[]} navData - an array of nav-data nodes, as detailed in [mktg-032](https://docs.google.com/document/d/1kYvbyd6njHFSscoE1dtDNHQ3U8IzaMdcjOS0jg87rHg)
 * @param {string} baseRoute - the base route where the navData will be rendered. For example, "docs".
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
 * @param {string} baseRoute - the base route where the navData will be rendered. For example, "docs".
 * @returns {Object[]} - an array of searchObjects to pass to Algolia. Must include an objectID property. See https://www.algolia.com/doc/api-reference/api-methods/add-objects/?client=javascript#examples.
 */
async function searchObjectsFromNavNode(node, baseRoute) {
  // If this is a node, build a search object
  if (node.path) {
    //  Fetch the MDX file content
    const fileString = node.filePath
      ? fs.readFileSync(path.join(process.cwd(), node.filePath), 'utf8')
      : node.remoteFile.fileString
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
