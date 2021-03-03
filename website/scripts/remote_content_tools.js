const fs = require('fs')
const path = require('path')
const dotenv = require('dotenv')

// Read in envs (need GITHUB_API_TOKEN from .env.local when running locally)
dotenv.config()
dotenv.config({ path: path.resolve(process.cwd(), '.env.local') })

const mergeRemotePlugins = require('../components/remote-plugin-docs/utils/merge-remote-plugins')
const fetchGithubFile = require('../components/remote-plugin-docs/utils/fetch-github-file')

async function writeRemoteContent(navDataFile, remotePluginsFile, contentDir) {
  const navData = readJson(navDataFile)
  const remotePlugins = readJson(remotePluginsFile)
  // Gather all remote plugins, merged into local navData
  const allNavData = await mergeRemotePlugins(remotePlugins, navData)
  // Traverse the navData tree, accumulating all remoteFile nodes,
  // each of which corresponds to an .mdx file in a remote plugin repo,
  // into a flattened array of NavLeafRemote navNodes
  const remoteNavLeaves = flattenRemoteNodes(allNavData)
  // TODO - as above
  // For each remote plugin file,
  // write it into the "local content" location to which
  // it corresponds, creating any parent folders as needed
  // const writtenFiles = await Promise.all( ??? )
  const contentDirPath = path.join(process.cwd(), contentDir)
  const remoteNavFiles = await Promise.all(
    remoteNavLeaves.map(async (node) => {
      const [err, fileContents] = await fetchGithubFile(node.remoteFile)
      if (err) throw new Error(err)
      const filePath = path.join(contentDirPath, `${node.path}.mdx`)
      return { filePath, fileContents }
    })
  )
  // Write out the files
  await Promise.all(
    remoteNavFiles.map(async (node) => {
      const { filePath, fileContents } = node
      //  Ensure the directory exists
      // (it might not, if the plugin has its own "tree" of content)
      const fileDir = path.dirname(filePath)
      fs.mkdirSync(fileDir, { recursive: true })
      //  Write out the file
      fs.writeFileSync(filePath, fileContents)
    })
  )
  return remoteNavFiles
}

// Clean up remote content files,
// which we temporarily wrote into place
async function cleanupRemoteContent(remoteFiles) {
  // For each remote plugin file,
  // delete the file
  await Promise.all(
    remoteFiles.map(async (node) => fs.unlinkSync(node.filePath))
  )
  // Delete any empty directories
  // with the local content folder
  // (assumed to be attributable to
  // nested remote docs content)
  await Promise.all(
    remoteFiles.map(async (node) => {
      const fileDir = path.dirname(node.filePath)
      if (fs.existsSync(fileDir)) {
        try {
          fs.rmdirSync(fileDir)
        } catch (err) {
          // We expect ENOTEMPTY errors, all other
          // errors are unexpected
          if (err.code !== 'ENOTEMPTY') throw new Error(err)
        }
      }
    })
  )
}

//
//
//

// Given navData, return a flattened
// array of all navNodes which are
// NavLeafRemote nodes
function flattenRemoteNodes(navData) {
  return navData.reduce((acc, navNode) => {
    // If it's a remote NavLeaf, then add it
    const isNavLeafRemote = !!navNode.remoteFile
    if (isNavLeafRemote) return acc.concat(navNode)
    // If it's a branch, then recurse
    const isNavBranch = !!navNode.routes
    if (isNavBranch) return acc.concat(flattenRemoteNodes(navNode.routes))
    // If it's anything else, carry on
    return acc
  }, [])
}

// Given a file path relative to
// the current working directory,
// read the file, parse it as json,
// and return the parsed data
function readJson(file) {
  const filePath = path.join(process.cwd(), file)
  const rawString = fs.readFileSync(filePath, 'utf-8')
  const jsonData = JSON.parse(rawString)
  return jsonData
}

module.exports = { cleanupRemoteContent, writeRemoteContent }
