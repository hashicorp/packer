require('dotenv').config()
const fs = require('fs')
const path = require('path')
const {
  indexContent,
  getDocsSearchObject,
} = require('@hashicorp/react-search/tools')
const resolveNavData = require('../components/remote-plugin-docs/utils/resolve-nav-data')
const fetchGithubFile = require('../components/remote-plugin-docs/utils/fetch-github-file')

// const NAV_DATA = 'data/docs-nav-data.json'
// const REMOTE_PLUGINS = 'data/docs-remote-plugins.json'
const CONTENT_DIR = 'content'

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
  // Get local files
  const projectRoot = process.cwd()
  // Get a single test object
  const contentDir = path.join(projectRoot, CONTENT_DIR)
  const fullPath = path.join(contentDir, 'docs', 'commands', 'build.mdx')
  const fileString = fs.readFileSync(fullPath, 'utf8')
  const urlPath = fullPath.replace(`${contentDir}/`, '').replace('.mdx', '')
  const testObject = await getDocsSearchObject(urlPath, fileString)
  console.log({ testObject })
  //
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

indexContent({ getSearchObjects })
