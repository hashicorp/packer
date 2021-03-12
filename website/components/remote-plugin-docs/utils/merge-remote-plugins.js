const path = require('path')
const fetchGithubFile = require('./fetch-github-file')

const COMPONENT_TYPES = [
  'builders',
  'datasources',
  'post-processors',
  'provisioners',
]

async function gatherRemotePlugins(pluginsData, navData, isDev = true) {
  const allPluginData = await Promise.all(
    pluginsData.map(async (pluginEntry) => {
      const componentEntries = await Promise.all(
        COMPONENT_TYPES.map(async (type) => {
          const routes = await gatherPluginBranch(pluginEntry, type)
          if (!routes) return false
          const isSingleLeaf =
            routes.length === 1 && typeof routes[0].path !== 'undefined'
          const navData = isSingleLeaf
            ? { ...routes[0], path: path.join(type, pluginEntry.path) }
            : { title: pluginEntry.title, routes }
          return { type, navData }
        })
      )
      const validComponents = componentEntries.filter(Boolean)
      if (validComponents.length === 0) {
        const errMsg = `Could not fetch any component documentation for remote plugin from ${pluginEntry.repo}. This may be a GitHub credential issue at build time, or it may be an issue with missing docs in the source repository. Please ensure you have a valid GITHUB_API_TOKEN set in .env.local at the root of the project.`
        if (isDev) {
          console.warn(errMsg)
        } else {
          throw new Error(errMsg)
        }
      }
      return validComponents
    })
  )

  const allPluginsByType = allPluginData.reduce((acc, pluginData) => {
    pluginData.forEach((p) => {
      const { type, navData } = p
      if (!acc[type]) acc[type] = []
      acc[type].push(navData)
    })
    return acc
  }, {})

  const navDataWithPlugins = navData.slice().map((n) => {
    // we only care about top-level NavBranch nodes
    if (!n.routes) return n
    // for each component type, check if this NavBranch
    // is the parent route for that type
    for (var i = 0; i < COMPONENT_TYPES.length; i++) {
      const type = COMPONENT_TYPES[i]
      const isTypeRoute = n.routes.filter((nn) => nn.path === type).length > 0
      if (isTypeRoute) {
        const pluginsOfType = allPluginsByType[type]
        if (!pluginsOfType || pluginsOfType.length == 0) return n
        // if this NavBranch is the parent route for the type,
        // then append all remote plugins of this type to the
        // NavBranch's child routes
        const routesWithPlugins = n.routes.slice().concat(pluginsOfType)
        // console.log(JSON.stringify(routesWithPlugins, null, 2))
        // Also, sort the child routes so the order is alphabetical
        routesWithPlugins.sort((a, b) => {
          // (exception: "Overview" comes first)
          if (a.title == 'Overview') return -1
          if (b.title === 'Overview') return 1
          // (exception: "Community-Supported" comes last)
          if (a.title == 'Community-Supported') return 1
          if (b.title === 'Community-Supported') return -1
          // (exception: "Custom" comes second-last)
          if (a.title == 'Custom') return 1
          if (b.title === 'Custom') return -1
          return a.title < b.title ? -1 : a.title > b.title ? 1 : 0
        })
        // return n
        return { ...n, routes: routesWithPlugins }
      }
    }
    return n
  })

  return navDataWithPlugins
}

async function gatherPluginBranch(pluginEntry, component) {
  const artifactDir = pluginEntry.artifactDir || '.docs-artifacts'
  const branch = pluginEntry.branch || 'main'
  const navDataFilePath = `${artifactDir}/${component}/nav-data.json`
  const [err, fileResult] = await fetchGithubFile({
    repo: pluginEntry.repo,
    branch,
    filePath: navDataFilePath,
  })
  // If one component errors, that's expected - we try all components.
  // We'll check one level up to see if ALL components fail.
  if (err) return false
  const navData = JSON.parse(fileResult)
  const withPrefixedPath = await prefixNavDataPath(
    navData,
    {
      repo: pluginEntry.repo,
      branch,
      componentArtifactsDir: path.join('.docs-artifacts', component),
    },
    path.join(component, pluginEntry.path)
  )
  // Add plugin tier
  // Parse the plugin tier
  const pluginOwner = pluginEntry.repo.split('/')[0]
  const pluginTier = pluginOwner === 'hashicorp' ? 'official' : 'community'
  const withPluginTier = addPluginTier(withPrefixedPath, pluginTier)
  //  Return the augmented navData
  return withPluginTier
}

function addPluginTier(navData, pluginTier) {
  return navData.slice().map((navNode) => {
    if (typeof navNode.path !== 'undefined') {
      return { ...navNode, pluginTier }
    }
    if (navNode.routes) {
      return { ...navNode, routes: addPluginTier(navNode.routes, pluginTier) }
    }
    return navNode
  })
}

async function prefixNavDataPath(
  navData,
  { repo, branch, componentArtifactsDir },
  parentPath
) {
  return await Promise.all(
    navData.slice().map(async (navNode) => {
      if (typeof navNode.path !== 'undefined') {
        const prefixedPath = path.join(parentPath, navNode.path)
        const remoteFile = {
          repo,
          branch,
          filePath: path.join(componentArtifactsDir, navNode.filePath),
        }
        const withPrefixedRoute = {
          ...navNode,
          path: prefixedPath,
          remoteFile: remoteFile,
        }
        delete withPrefixedRoute.filePath
        return withPrefixedRoute
      }
      if (navNode.routes) {
        const prefixedRoutes = await prefixNavDataPath(
          navNode.routes,
          { repo, branch, componentArtifactsDir },
          parentPath
        )
        const withPrefixedRoutes = { ...navNode, routes: prefixedRoutes }
        return withPrefixedRoutes
      }
      return navNode
    })
  )
}

module.exports = gatherRemotePlugins
