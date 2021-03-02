import fs from 'fs'
import path from 'path'
import {
  getNodeFromPath,
  getPathsFromNavData,
  validateNavData,
} from '@hashicorp/react-docs-page/server'
import renderPageMdx from '@hashicorp/react-docs-page/render-page-mdx'
import fetchGithubFile from './utils/fetch-github-file'
import mergeRemotePlugins from './utils/merge-remote-plugins'

const IS_DEV = process.env.VERCEL_ENV !== 'production'

async function generateStaticPaths(navDataFile, contentDir, options = {}) {
  const navData = await resolveNavData(navDataFile, contentDir, options)
  const paths = await getPathsFromNavData(navData)
  return paths
}

async function generateStaticProps(
  navDataFile,
  localContentDir,
  params,
  { productName, remotePluginsFile, additionalComponents } = {}
) {
  const navData = await resolveNavData(navDataFile, localContentDir, {
    remotePluginsFile,
  })
  const pathToMatch = params.page ? params.page.join('/') : ''
  const navNode = getNodeFromPath(pathToMatch, navData, localContentDir)
  const { filePath, remoteFile, pluginTier } = navNode
  //  Fetch the MDX file content
  const [err, mdxString] = filePath
    ? //  Read local content from the filesystem
      [null, fs.readFileSync(path.join(process.cwd(), filePath), 'utf8')]
    : // Fetch remote content using GitHub's API
      await fetchGithubFile(remoteFile)
  if (err) throw new Error(err)
  // For plugin pages, prefix the MDX content with a
  // label that reflects the plugin tier
  // (current options are "Official" or "Community")
  function mdxContentHook(mdxContent) {
    if (pluginTier) {
      const tierMdx = `<PluginTierLabel tier="${pluginTier}" />\n\n`
      mdxContent = tierMdx + mdxContent
    }
    return mdxContent
  }
  const { mdxSource, frontMatter } = await renderPageMdx(mdxString, {
    additionalComponents,
    productName,
    mdxContentHook,
  })
  // Build the currentPath from page parameters
  const currentPath = !params.page ? '' : params.page.join('/')
  // In development, set a flag if there is no GITHUB_API_TOKEN,
  // as this means dev is seeing only local content, and we want to flag that
  const isDevMissingRemotePlugins = IS_DEV && !process.env.GITHUB_API_TOKEN
  return {
    currentPath,
    frontMatter,
    isDevMissingRemotePlugins,
    mdxSource,
    mdxString,
    navData,
    navNode,
  }
}

async function resolveNavData(navDataFile, localContentDir, options = {}) {
  const { remotePluginsFile } = options
  // Read in files
  const navDataPath = path.join(process.cwd(), navDataFile)
  const navData = JSON.parse(fs.readFileSync(navDataPath, 'utf8'))
  const remotePluginsPath = path.join(process.cwd(), remotePluginsFile)
  const remotePlugins = JSON.parse(fs.readFileSync(remotePluginsPath, 'utf-8'))
  // Resolve plugins, this yields branches with NavLeafRemote nodes
  const withPlugins = await mergeRemotePlugins(remotePlugins, navData, IS_DEV)
  // Resolve local filePaths for NavLeaf nodes
  const withFilePaths = await validateNavData(withPlugins, localContentDir)
  // Return the nav data with:
  // 1. Plugins merged, transformed into navData structures with NavLeafRemote nodes
  // 2. filePaths added to all local NavLeaf nodes
  return withFilePaths
}

export default { generateStaticPaths, generateStaticProps }
export { generateStaticPaths, generateStaticProps }
