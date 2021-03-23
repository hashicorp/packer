import { productName, productSlug } from 'data/metadata'
import DocsPage from '@hashicorp/react-docs-page'
import PluginTierLabel from 'components/plugin-tier-label'
import DevAlert from 'components/dev-alert'
// Imports below are only used server-side
import {
  generateStaticPaths,
  generateStaticProps,
} from 'components/remote-plugin-docs/server'

//  Configure the docs path
const BASE_ROUTE = 'docs'
const NAV_DATA = 'data/docs-nav-data.json'
const CONTENT_DIR = 'content/docs'
const PRODUCT = { name: productName, slug: productSlug }
// add remote plugin docs loading
const OPTIONS = {
  remotePluginsFile: 'data/docs-remote-plugins.json',
  additionalComponents: { PluginTierLabel },
  mainBranch: 'master',
}

function DocsLayout({ isDevMissingRemotePlugins, ...props }) {
  return (
    <>
      {isDevMissingRemotePlugins ? (
        <DevAlert>
          <strong className="g-type-label-strong">
            Note for local development
          </strong>
          <p>
            <span role="img" aria-label="Alert: ">
              🚨
            </span>{' '}
            <strong>This preview is missing plugin docs</strong> pulled from
            remote repos.
          </p>

          <p>
            <span role="img" aria-label="Fix: ">
              🛠
            </span>{' '}
            <strong>To preview docs pulled from plugin repos</strong>, please
            include a <code>GITHUB_API_TOKEN</code> in{' '}
            <code>website/.env.local</code>.
          </p>
        </DevAlert>
      ) : null}
      <DocsPage
        additionalComponents={OPTIONS.additionalComponents}
        baseRoute={BASE_ROUTE}
        product={PRODUCT}
        staticProps={props}
      />
    </>
  )
}

export async function getStaticPaths() {
  const paths = await generateStaticPaths(NAV_DATA, CONTENT_DIR, OPTIONS)
  return { paths, fallback: false }
}

export async function getStaticProps({ params }) {
  const props = await generateStaticProps(
    NAV_DATA,
    CONTENT_DIR,
    params,
    PRODUCT,
    OPTIONS
  )
  return { props }
}

export default DocsLayout
