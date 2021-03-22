import fs from 'fs'
import path from 'path'
import {
  getNodeFromPath,
  getPathsFromNavData,
} from '@hashicorp/react-docs-page/server'
import renderPageMdx from '@hashicorp/react-docs-page/render-page-mdx'
import resolveNavData from './utils/resolve-nav-data'

async function generateStaticPaths(navDataFile, contentDir, options = {}) {
  const navData = await resolveNavData(navDataFile, contentDir, options)
  const paths = await getPathsFromNavData(navData)
  return paths
}

async function generateStaticProps(
  navDataFile,
  localContentDir,
  params,
  product,
  { remotePluginsFile, additionalComponents, mainBranch = 'main' } = {}
) {
  const navData = await resolveNavData(navDataFile, localContentDir, {
    remotePluginsFile,
  })
  const pathToMatch = params.page ? params.page.join('/') : ''
  const navNode = getNodeFromPath(pathToMatch, navData, localContentDir)
  const { filePath, remoteFile, pluginTier } = navNode
  //  Fetch the MDX file content
  const mdxString = remoteFile
    ? remoteFile.fileString
    : fs.readFileSync(path.join(process.cwd(), filePath), 'utf8')
  // Construct the githubFileUrl, used for "Edit this page" link
  // Note: we expect remote files, such as those used to render plugin docs,
  // to have a sourceUrl defined, that points to the file we built from
  const githubFileUrl = remoteFile
    ? remoteFile.sourceUrl
    : `https://github.com/hashicorp/${product.slug}/blob/${mainBranch}/website/${filePath}`
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
    productName: product.name,
    mdxContentHook,
  })
  // Build the currentPath from page parameters
  const currentPath = !params.page ? '' : params.page.join('/')

  return {
    currentPath,
    frontMatter,
    mdxSource,
    mdxString,
    githubFileUrl,
    navData,
    navNode,
  }
}

export default { generateStaticPaths, generateStaticProps }
export { generateStaticPaths, generateStaticProps }
