import { productName, productSlug } from 'data/metadata'
import DocsPage from '@hashicorp/react-docs-page'
// Imports below are only used server-side
import { getStaticGenerationFunctions } from '@hashicorp/react-docs-page/server'

//  Configure the docs path
const baseRoute = 'guides'
const navDataFile = 'data/guides-nav-data.json'
const localContentDir = 'content/guides'
const mainBranch = 'master'
const product = { name: productName, slug: productSlug }

export default function GuidesLayout(props) {
  return (
    <DocsPage baseRoute={baseRoute} product={product} staticProps={props} />
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
