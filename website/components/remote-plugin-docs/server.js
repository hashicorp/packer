import fs from 'fs'
import path from 'path'
import {
  getNodeFromPath,
  getPathsFromNavData,
} from '@hashicorp/react-docs-page/server'
import renderPageMdx from '@hashicorp/react-docs-page/render-page-mdx'
import fetchGithubFile from './utils/fetch-github-file'
import resolveNavData from './utils/resolve-nav-data'

const IS_DEV = process.env.VERCEL_ENV !== 'production'

async function generateStaticPaths(navDataFile, contentDir, options = {}) {
  const navData = await resolveNavData(navDataFile, contentDir, {
    ...options,
    isDev: IS_DEV,
  })
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
    isDev: IS_DEV,
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
  // Construct the githubFileUrl, used for "Edit this page" link
  // Note: for custom ".docs-artifacts" directories, the "Edit this page"
  // link will lead to the artifact file rather than the "docs" source file
  const githubFileUrl = filePath
    ? `https://github.com/hashicorp/${product.slug}/blob/${mainBranch}/website/${filePath}`
    : `https://github.com/${remoteFile.repo}/blob/${
        remoteFile.branch
      }/${remoteFile.filePath.replace('\b.docs-artifacts', 'docs')}`
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

  // In development, set a flag if there is no GITHUB_API_TOKEN,
  // as this means dev is seeing only local content, and we want to flag that
  const isDevMissingRemotePlugins = IS_DEV && !process.env.GITHUB_API_TOKEN
  return {
    currentPath,
    frontMatter,
    isDevMissingRemotePlugins,
    mdxSource,
    mdxString,
    githubFileUrl,
    navData,
    navNode,
  }
}

export default { generateStaticPaths, generateStaticProps }
export { generateStaticPaths, generateStaticProps }
