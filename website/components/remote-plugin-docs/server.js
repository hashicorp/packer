import fs from 'fs'
import path from 'path'
import {
  getNodeFromPath,
  getPathsFromNavData,
} from '@hashicorp/react-docs-page/server'
import renderPageMdx from '@hashicorp/react-docs-page/render-page-mdx'
import resolveNavDataWithRemotePlugins from './utils/resolve-nav-data'
import fetchLatestReleaseTag from './utils/fetch-latest-release-tag'

async function generateStaticPaths({ navDataFile, remotePluginsFile }) {
  const navData = await resolveNavDataWithRemotePlugins(navDataFile, {
    remotePluginsFile,
  })
  const paths = await getPathsFromNavData(navData)
  return paths
}

async function generateStaticProps({
  additionalComponents,
  localContentDir,
  mainBranch = 'main',
  navDataFile,
  params,
  product,
  remotePluginsFile,
}) {
  // Build the currentPath from page parameters
  const currentPath = params.page ? params.page.join('/') : ''
  // Resolve navData, including the possibility that this
  // page is a remote plugin docs, in which case we'll provide
  // the MDX fileString in the resolved navData
  const navData = await resolveNavDataWithRemotePlugins(navDataFile, {
    remotePluginsFile,
    currentPath,
  })
  const navNode = getNodeFromPath(currentPath, navData, localContentDir)
  const { filePath, remoteFile, pluginData } = navNode
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
  // If this is a plugin, and if
  // the version has been specified as "latest",
  // determine the tag this corresponds to, so that
  // we can show this explicit version number in docs
  const latestReleaseTag =
    pluginData?.version === 'latest'
      ? await fetchLatestReleaseTag(pluginData.repo)
      : pluginData?.version
  // For plugin pages, prefix the MDX content with a
  // label that reflects the plugin tier
  // (current options are "Official" or "Community")
  // and display whether the plugin is "HCP Packer Ready".
  // Also add a badge to show the latest version
  function mdxContentHook(mdxContent) {
    const badgesMdx = []
    // Add a badge for the plugin tier
    if (pluginData?.tier) {
      badgesMdx.push(`<PluginBadge type="${pluginData.tier}" />`)
    }
    // Add a badge if the plugin is "HCP Packer Ready"
    if (pluginData?.isHcpPackerReady) {
      badgesMdx.push(`<PluginBadge type="hcp_packer_ready" />`)
    }
    // If the plugin is archived, add an "Archived" badge
    if (pluginData?.archived == true) {
      badgesMdx.push(`<PluginBadge type="archived" />`)
    }
    // Add badge showing the latest release version number,
    // and link this badge to the latest release
    if (latestReleaseTag) {
      const href = `https://github.com/${pluginData.repo}/releases/tag/${latestReleaseTag}`
      badgesMdx.push(
        `<Badge href="${href}" label="${latestReleaseTag}" theme="light-gray"/>`
      )
    }
    // If we have badges to add, inject them into the MDX
    if (badgesMdx.length > 0) {
      const badgeChildrenMdx = badgesMdx.join('')
      const badgesHeaderMdx = `<BadgesHeader>${badgeChildrenMdx}</BadgesHeader>`
      mdxContent = badgesHeaderMdx + '\n\n' + mdxContent
    }
    return mdxContent
  }
  const { mdxSource, frontMatter } = await renderPageMdx(mdxString, {
    additionalComponents,
    productName: product.name,
    mdxContentHook,
  })

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

export { generateStaticPaths, generateStaticProps }
