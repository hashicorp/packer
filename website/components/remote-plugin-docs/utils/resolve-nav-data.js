const fs = require('fs')
const path = require('path')
const grayMatter = require('gray-matter')
const fetchPluginDocs = require('./fetch-plugin-docs')
const fetchDevPluginDocs = require('./fetch-dev-plugin-docs')
const validateFilePaths = require('@hashicorp/react-docs-sidenav/utils/validate-file-paths')
const validateRouteStructure = require('@hashicorp/react-docs-sidenav/utils/validate-route-structure')

/**
 * Resolves nav-data from file, including optional
 * resolution of remote plugin docs entries
 *
 * @param {string} navDataFile path to the nav-data.json file, relative to the cwd. Example: "data/docs-nav-data.json".
 * @param {string} localContentDir path to the content root, relative to the cwd. Example: "content/docs".
 * @param {object} options optional configuration object
 * @param {string} options.remotePluginsFile path to a remote-plugins.json file, relative to the cwd. Example: "data/docs-remote-plugins.json".
 * @returns {array} the resolved navData. This includes NavBranch nodes pulled from remote plugin repositories, as well as filePath properties on all local NavLeaf nodes, and remoteFile properties on all NavLeafRemote nodes.
 */
async function resolveNavData(navDataFile, localContentDir, options = {}) {
  const { remotePluginsFile, currentPath } = options
  // Read in files
  const navDataPath = path.join(process.cwd(), navDataFile)
  const navData = JSON.parse(fs.readFileSync(navDataPath, 'utf8'))
  // Fetch remote plugin docs, if applicable
  let withPlugins = navData
  if (remotePluginsFile) {
    // Resolve plugins, this yields branches with NavLeafRemote nodes
    withPlugins = await mergeRemotePlugins(
      remotePluginsFile,
      navData,
      currentPath
    )
  }
  // Resolve local filePaths for NavLeaf nodes
  const withFilePaths = await validateFilePaths(withPlugins, localContentDir)
  validateRouteStructure(withFilePaths)
  // Return the nav data with:
  // 1. Plugins merged, transformed into navData structures with NavLeafRemote nodes
  // 2. filePaths added to all local NavLeaf nodes
  return withFilePaths
}

// Given a remote plugins config file, and the full tree of docs navData which
// contains top-level branch routes that match plugin component types,
// fetch and parse all remote plugin docs, merge them into the
// broader tree of docs navData, and return the docs navData
// with the merged plugin docs
async function mergeRemotePlugins(remotePluginsFile, navData, currentPath) {
  // Read in and parse the plugin configuration JSON
  const remotePluginsPath = path.join(process.cwd(), remotePluginsFile)
  const pluginEntries = JSON.parse(fs.readFileSync(remotePluginsPath, 'utf-8'))
  // Add navData for each plugin's component.
  // Note that leaf nodes include a remoteFile property object with the full MDX fileString
  const pluginEntriesWithDocs = await Promise.all(
    pluginEntries.map(
      async (entry) => await resolvePluginEntryDocs(entry, currentPath)
    )
  )
  // group navData by component type, to prepare to merge plugin docs
  // into the broader tree of navData.
  const pluginDocsByComponent = pluginEntriesWithDocs.reduce(
    (acc, pluginEntry) => {
      const { components } = pluginEntry
      Object.keys(components).forEach((type) => {
        const navData = components[type]
        if (!navData) return
        if (!acc[type]) acc[type] = []
        acc[type].push(navData[0])
      })
      return acc
    },
    {}
  )
  // merge plugin docs, by plugin component type,
  // into the corresponding top-level component NavBranch
  const navDataWithPlugins = navData.slice().map((n) => {
    // we only care about top-level NavBranch nodes
    if (!n.routes) return n
    // for each component type, check if this NavBranch
    // is the parent route for that type
    const componentTypes = Object.keys(pluginDocsByComponent)
    let typeMatch = false
    for (var i = 0; i < componentTypes.length; i++) {
      const componentType = componentTypes[i]
      const routeMatches = n.routes.filter((r) => r.path === componentType)
      if (routeMatches.length > 0) {
        typeMatch = componentType
        break
      }
    }
    // if this NavBranch does not match a component type slug,
    // then return it unmodified
    if (!typeMatch) return n
    // if there are no matching remote plugin components,
    // then return the navBranch unmodified
    const pluginsOfType = pluginDocsByComponent[typeMatch]
    if (!pluginsOfType || pluginsOfType.length == 0) return n
    // if this NavBranch is the parent route for the type,
    // then append all remote plugins of this type to the
    // NavBranch's child routes
    const routesWithPlugins = n.routes.slice().concat(pluginsOfType)
    // console.log(JSON.stringify(routesWithPlugins, null, 2))
    // Also, sort the child routes so the order is alphabetical
    routesWithPlugins.sort((a, b) => {
      // ensure casing does not affect ordering
      const aTitle = a.title.toLowerCase()
      const bTitle = b.title.toLowerCase()
      // (exception: "Overview" comes first)
      if (aTitle === 'overview') return -1
      if (bTitle === 'overview') return 1
      // (exception: "Community-Supported" comes last)
      if (aTitle === 'community-supported') return 1
      if (bTitle === 'community-supported') return -1
      // (exception: "Custom" comes second-last)
      if (aTitle === 'custom') return 1
      if (bTitle === 'custom') return -1
      return aTitle < bTitle ? -1 : aTitle > bTitle ? 1 : 0
    })
    // return n
    return { ...n, routes: routesWithPlugins }
  })
  // return the merged navData, which now includes special NavLeaf nodes
  // for plugin docs with remoteFile properties
  return navDataWithPlugins
}

// Fetch remote plugin docs .mdx files, and
// transform each plugin's array of .mdx files into navData.
// Organize this navData by component, add it to the plugin config entry,
// and return the modified entry.
//
// Note that navData leaf nodes have a special remoteFile property,
// which contains { filePath, fileString } data for the remote
// plugin doc .mdx file
async function resolvePluginEntryDocs(pluginConfigEntry, currentPath) {
  const {
    title,
    path: slug,
    repo,
    version,
    pluginTier,
    sourceBranch = 'main',
    zipFile = '',
  } = pluginConfigEntry
  var docsMdxFiles
  if (zipFile !== '') {
    docsMdxFiles = await fetchDevPluginDocs(zipFile)
  } else {
    docsMdxFiles = await fetchPluginDocs({ repo, tag: version })
  }
  // We construct a special kind of "NavLeaf" node, with a remoteFile property,
  // consisting of a { filePath, fileString, sourceUrl }, where:
  // - filePath is the path to the source file in the source repo
  // - fileString is a string representing the file source
  // - sourceUrl is a link to the original file in the source repo
  // We also add a pluginTier attribute
  const navNodes = docsMdxFiles.map((mdxFile) => {
    const { filePath, fileString } = mdxFile
    // Process into a NavLeaf, with a remoteFile attribute
    const dirs = path.dirname(filePath).split('/')
    const dirUrl = dirs.slice(2).join('/')
    const basename = path.basename(filePath, path.extname(filePath))
    // build urlPath
    // note that this will be prefixed to get to our final path
    const isIndexFile = basename === 'index'
    const urlPath = isIndexFile ? dirUrl : path.join(dirUrl, basename)
    // parse title, either from frontmatter or file name
    const { data: frontmatter } = grayMatter(fileString)
    const { nav_title, sidebar_title } = frontmatter
    const title = nav_title || sidebar_title || basename
    // construct sourceUrl (used for "Edit this page" link)
    const sourceUrl = `https://github.com/${repo}/blob/${sourceBranch}/${filePath}`
    // determine pluginTier
    const pluginOwner = repo.split('/')[0]
    const parsedPluginTier =
      pluginTier || (pluginOwner === 'hashicorp' ? 'official' : 'community')
    // Construct and return a NavLeafRemote node
    return {
      title,
      path: urlPath,
      remoteFile: { filePath, fileString, sourceUrl },
      pluginTier: parsedPluginTier,
    }
  })
  //
  navNodes.sort((a, b) => {
    // ensure casing does not affect ordering
    const aTitle = a.title.toLowerCase()
    const bTitle = b.title.toLowerCase()
    // (exception: "Overview" comes first)
    if (aTitle === 'overview') return -1
    if (bTitle === 'overview') return 1
    return aTitle < bTitle ? -1 : aTitle > bTitle ? 1 : 0
  })
  //
  const navNodesByComponent = navNodes.reduce((acc, navLeaf) => {
    const componentType = navLeaf.remoteFile.filePath.split('/')[1]
    if (!acc[componentType]) acc[componentType] = []
    acc[componentType].push(navLeaf)
    return acc
  }, {})
  //
  const components = Object.keys(navNodesByComponent).map((type) => {
    // Plugins many not contain every component type,
    // we return null if this is the case
    const rawNavNodes = navNodesByComponent[type]
    if (!rawNavNodes) return null
    // Avoid unnecessary nesting if there's only a single doc file
    const navData = normalizeNavNodes(title, rawNavNodes)
    // Prefix paths to fit into broader docs nav-data
    const pathPrefix = path.join(type, slug)
    const withPrefixedPaths = visitNavLeaves(navData, (n) => {
      const prefixedPath = path.join(pathPrefix, n.path)
      return { ...n, path: prefixedPath }
    })
    // If currentPath is provided, then remove the remoteFile
    // from all nodes except the currentPath. This ensures we deliver
    // only a single fileString in our getStaticProps JSON.
    // Without this optimization, we would send all fileStrings
    // for all NavLeafRemote nodes
    const withOptimizedFileStrings = visitNavLeaves(withPrefixedPaths, (n) => {
      if (!n.remoteFile) return n
      const noCurrentPath = typeof currentPath === 'undefined'
      const isPathMatch = currentPath === n.path
      if (noCurrentPath || isPathMatch) return n
      const { filePath } = n.remoteFile
      return { ...n, remoteFile: { filePath } }
    })
    // Return the component, with processed navData
    return { type, navData: withOptimizedFileStrings }
  })
  const componentsObj = components.reduce((acc, component) => {
    if (!component) return acc
    acc[component.type] = component.navData
    return acc
  }, {})
  return { ...pluginConfigEntry, components: componentsObj }
}

// For components with a single doc file, transform so that
// a single leaf node renders, rather than a nav branch
function normalizeNavNodes(pluginName, routes) {
  const isSingleLeaf =
    routes.length === 1 && typeof routes[0].path !== 'undefined'
  const navData = isSingleLeaf
    ? [{ ...routes[0], path: '' }]
    : [{ title: pluginName, routes }]
  return navData
}

// Traverse a clone of the given navData,
// modifying any NavLeaf nodes with the provided visitFn
function visitNavLeaves(navData, visitFn) {
  return navData.slice().map((navNode) => {
    if (typeof navNode.path !== 'undefined') {
      return visitFn(navNode)
    }
    if (navNode.routes) {
      return { ...navNode, routes: visitNavLeaves(navNode.routes, visitFn) }
    }
    return navNode
  })
}

module.exports = resolveNavData
