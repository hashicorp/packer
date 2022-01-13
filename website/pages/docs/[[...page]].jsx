import { productName, productSlug } from 'data/metadata'
import DocsPage from '@hashicorp/react-docs-page'
import Badge from 'components/badge'
import BadgesHeader from 'components/badges-header'
import PluginBadge from 'components/plugin-badge'
import Checklist from 'components/checklist'
// Imports below are only used server-side
import { getStaticGenerationFunctions } from '@hashicorp/react-docs-page/server'

//  Configure the docs path and remote plugin docs loading
const additionalComponents = { Badge, BadgesHeader, PluginBadge, Checklist }
const baseRoute = 'docs'
const localContentDir = 'content/docs'
const mainBranch = 'master'
const navDataFile = 'data/docs-nav-data.json'
const product = { name: productName, slug: productSlug }

export default function DocsLayout({ isDevMissingRemotePlugins, ...props }) {
  return (
    <DocsPage
      additionalComponents={additionalComponents}
      baseRoute={baseRoute}
      product={product}
      staticProps={props}
    />
  )
}

const { getStaticPaths, getStaticProps } = getStaticGenerationFunctions(
  process.env.ENABLE_VERSIONED_DOCS === 'true'
    ? {
        strategy: 'remote',
        basePath: baseRoute,
        fallback: 'blocking',
        revalidate: 360, // 1 hour
        product: productSlug,
        mainBranch: mainBranch,
      }
    : {
        strategy: 'fs',
        localContentDir: localContentDir,
        navDataFile: navDataFile,
        product: productSlug,
        revalidate: false,
        mainBranch: mainBranch,
      }
)

export { getStaticPaths, getStaticProps }
