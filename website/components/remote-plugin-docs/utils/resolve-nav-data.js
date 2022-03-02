const fs = require('fs')
const path = require('path')
const grayMatter = require('gray-matter')
const fetchPluginDocs = require('./fetch-plugin-docs')
const fetchDevPluginDocs = require('./fetch-dev-plugin-docs')

/**
 * Resolves nav-data from file with
 * resolution of remote plugin docs entries
 *
 * @param {string} navDataFile path to the nav-data.json file, relative to the cwd. Example: "data/docs-nav-data.json".
 * @param {object} options optional configuration object
 * @param {string} options.remotePluginsFile path to a remote-plugins.json file, relative to the cwd. Example: "data/docs-remote-plugins.json".
 * @returns {Promise<array>} the resolved navData. This includes NavBranch nodes pulled from remote plugin repositories, as well as filePath properties on all local NavLeaf nodes, and remoteFile properties on all NavLeafRemote nodes.
 */
async function resolveNavDataWithRemotePlugins(navDataFile, options = {}) {
  const { remotePluginsFile, currentPath } = options
  const navDataPath = path.join(process.cwd(), navDataFile)
  let navData = JSON.parse(fs.readFileSync(navDataPath, 'utf8'))
  return await appendRemotePluginsNavData(
    remotePluginsFile,
    navData,
    currentPath
  )
}

async function appendRemotePluginsNavData(
  remotePluginsFile,
  navData,
  currentPath
) {
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

  const titleMap = {
    builders: 'Builders',
    provisioners: 'Provisioners',
    'post-processors': 'Post-Processors',
    datasources: 'Data Sources',
  }

  return navData.concat(
    pluginEntriesWithDocs.map((entry) => {
      return {
        title: entry.title,
        routes: Object.entries(entry.components).map(
          ([type, componentList]) => {
            return {
              title: titleMap[type],
              // Flat map to avoid ┐
              // > Proxmox         │
              //   > Builders      │
              //     > Proxmox <---┘
              //       > Overview
              //       > Clone
              //       > ISO
              routes: componentList.flatMap((c) => {
                if ('path' in c) {
                  return c
                } else if ('routes' in c) {
                  return c.routes
                }
              }),
            }
          }
        ),
      }
    })
  )
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
    archived = false,
    isHcpPackerReady = false,
    sourceBranch = 'main',
    zipFile = '',
  } = pluginConfigEntry
  // Determine the pluginTier, which can be set manually,
  // or will be automatically set based on repo ownership
  const pluginOwner = repo.split('/')[0]
  const parsedPluginTier =
    pluginTier || (pluginOwner === 'hashicorp' ? 'official' : 'community')
  // Fetch the MDX files for the plugin entry
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
  // We also add pluginData, which is used to add badges
  // such as the plugin's tier when rendering the page.
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
    // Construct and return a NavLeafRemote node
    return {
      title,
      path: urlPath,
      remoteFile: { filePath, fileString, sourceUrl },
      pluginData: {
        repo,
        tier: parsedPluginTier,
        isHcpPackerReady,
        version,
        archived,
      },
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

module.exports = resolveNavDataWithRemotePlugins
