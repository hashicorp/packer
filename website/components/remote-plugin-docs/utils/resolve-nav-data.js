const fs = require('fs')
const path = require('path')
const mergeRemotePlugins = require('./merge-remote-plugins')
const validateFilePaths = require('@hashicorp/react-docs-sidenav/utils/validate-file-paths')
const validateRouteStructure = require('@hashicorp/react-docs-sidenav/utils/validate-route-structure')

/**
 * Resolves nav-data from file, including optional
 * resolution of remote plugin docs entries
 *
 * @param {string} navDataFile path to the nav-data.json file, relative to the cwd. Example: "data/docs-nav-data.json".
 * @param {string} localContentDir path to the content root, relative to the cwd. Example: "content/docs".
 * @param {object} options optional configuration object
 * @param {string} options.isDev if true, then will NOT throw errors if remote fetches fail
 * @param {string} options.remotePluginsFile path to a remote-plugins.json file, relative to the cwd. Example: "data/docs-remote-plugins.json".
 * @returns {object} the resolved navData. This includes NavBranch nodes pulled from remote plugin repositories, as well as filePath properties on all local NavLeaf nodes, and remoteFile properties on all remote NavLeaf nodes.
 */
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
