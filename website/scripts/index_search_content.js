const fs = require('fs')
const path = require('path')
const {
  indexContent,
  getDocsSearchObject,
} = require('@hashicorp/react-search/tools')
const resolveNavData = require('../components/remote-plugin-docs/utils/resolve-nav-data')
const fetchGithubFile = require('../components/remote-plugin-docs/utils/fetch-github-file')
const dotenv = require('dotenv')

// Read in envs (need GITHUB_API_TOKEN from .env.local when running locally)
dotenv.config()
dotenv.config({ path: path.resolve(process.cwd(), '.env.local') })

async function getSearchObjects() {
  // Resolve /docs, /guides, and /intro nav data,
  // which corresponds to all the content we will
  // actually render (this avoids indexing non-rendered content, and partials)
  // `docs`
  const docsNav = await resolveNavData(
    'data/docs-nav-data.json',
    'content/docs',
    { remotePluginsFile: 'data/docs-remote-plugins.json' }
  )
  const docsObjects = await searchObjectsFromNavData(docsNav, 'docs')
  // `guides`
  const guidesNav = await resolveNavData(
    'data/guides-nav-data.json',
    'content/guides'
  )
  const guidesObjects = await searchObjectsFromNavData(guidesNav, 'guides')
  // `intro`
  const introNav = await resolveNavData(
    'data/intro-nav-data.json',
    'content/intro'
  )
  const introObjects = await searchObjectsFromNavData(introNav, 'intro')
  // Collect all search objects
  const searchObjects = [].concat(docsObjects, guidesObjects, introObjects)
  return searchObjects
}

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

async function searchObjectsFromNavNode(node, baseRoute) {
  // If this is a node, build a search object
  if (node.path) {
    //  Fetch the MDX file content
    const [err, fileString] = node.filePath
      ? //  Read local content from the filesystem
        [null, fs.readFileSync(path.join(process.cwd(), node.filePath), 'utf8')]
      : // Fetch remote content using GitHub's API
        await fetchGithubFile(node.remoteFile)
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

// Run indexing
indexContent({ getSearchObjects })
