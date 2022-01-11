import { productName, productSlug } from 'data/metadata'
import DocsPage from '@hashicorp/react-docs-page'
import Badge from 'components/badge'
import BadgesHeader from 'components/badges-header'
import PluginBadge from 'components/plugin-badge'
import Checklist from 'components/checklist'
// Imports below are only used server-side
import {
  generateStaticPaths,
  generateStaticProps,
} from '@hashicorp/react-docs-page/server'

//  Configure the docs path and remote plugin docs loading
const additionalComponents = { Badge, BadgesHeader, PluginBadge, Checklist }
const baseRoute = 'docs'
const localContentDir = 'content/docs'
const mainBranch = 'master'
const navDataFile = 'data/docs-nav-data.json'
const product = { name: productName, slug: productSlug }

function DocsLayout({ isDevMissingRemotePlugins, ...props }) {
  return (
    <DocsPage
      additionalComponents={additionalComponents}
      baseRoute={baseRoute}
      product={product}
      staticProps={props}
    />
  )
}

export async function getStaticPaths() {
  const paths = await generateStaticPaths({
    localContentDir,
    navDataFile,
  })
  return { paths, fallback: false }
}

export async function getStaticProps({ params }) {
  const props = await generateStaticProps({
    additionalComponents,
    localContentDir,
    mainBranch,
    navDataFile,
    params,
    product,
  })
  return { props }
}

export default DocsLayout
