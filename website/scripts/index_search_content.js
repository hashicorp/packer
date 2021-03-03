const {
  cleanupRemoteContent,
  writeRemoteContent,
} = require('./remote_content_tools')
const { indexDocsContent } = require('@hashicorp/react-search/tools')

const NAV_DATA = 'data/docs-nav-data.json'
const REMOTE_PLUGINS = 'data/docs-remote-plugins.json'
const CONTENT_DIR = 'content/docs'

indexAllContent()

async function indexAllContent() {
  // Temporarily write remote files to their place in the
  // docs content tree, so they're recognized
  // and indexed in indexDocsContent()
  const remoteFiles = await writeRemoteContent(
    NAV_DATA,
    REMOTE_PLUGINS,
    CONTENT_DIR
  )
  // Run the standard indexDocsContent() function
  await indexDocsContent()
  // Delete temporary remote files
  await cleanupRemoteContent(remoteFiles)
}
