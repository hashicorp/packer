const fs = require('fs')
const path = require('path')
const mergeRemotePlugins = require('./merge-remote-plugins')
const validateFilePaths = require('@hashicorp/react-docs-sidenav/utils/validate-file-paths')
const validateRouteStructure = require('@hashicorp/react-docs-sidenav/utils/validate-route-structure')

async function resolveNavData(navDataFile, localContentDir, options = {}) {
  const { remotePluginsFile, isDev } = options
  // Read in files
  const navDataPath = path.join(process.cwd(), navDataFile)
  const navData = JSON.parse(fs.readFileSync(navDataPath, 'utf8'))
  // Fetch remote plugin docs, if applicable
  let withPlugins = navData
  if (remotePluginsFile) {
    const remotePluginsPath = path.join(process.cwd(), remotePluginsFile)
    const remotePlugins = JSON.parse(
      fs.readFileSync(remotePluginsPath, 'utf-8')
    )
    // Resolve plugins, this yields branches with NavLeafRemote nodes
    withPlugins = await mergeRemotePlugins(remotePlugins, navData, isDev)
  }
  // Resolve local filePaths for NavLeaf nodes
  const withFilePaths = await validateFilePaths(withPlugins, localContentDir)
  validateRouteStructure(withFilePaths)
  // Return the nav data with:
  // 1. Plugins merged, transformed into navData structures with NavLeafRemote nodes
  // 2. filePaths added to all local NavLeaf nodes
  return withFilePaths
}

module.exports = resolveNavData
