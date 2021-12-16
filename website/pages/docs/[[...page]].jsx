import { productName, productSlug } from 'data/metadata'
import DocsPage from '@hashicorp/react-docs-page'
import Badge from 'components/badge'
import BadgesHeader from 'components/badges-header'
import PluginBadge from 'components/plugin-badge'
import DevAlert from 'components/dev-alert'
import Checklist from 'components/checklist'
// Imports below are only used server-side
import {
  generateStaticPaths,
  generateStaticProps,
} from 'components/remote-plugin-docs/server'

//  Configure the docs path and remote plugin docs loading
const additionalComponents = { Badge, BadgesHeader, PluginBadge, Checklist }
const baseRoute = 'docs'
const localContentDir = 'content/docs'
const mainBranch = 'master'
const navDataFile = 'data/docs-nav-data.json'
const product = { name: productName, slug: productSlug }
const remotePluginsFile = 'data/docs-remote-plugins.json'

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
        additionalComponents={additionalComponents}
        baseRoute={baseRoute}
        product={product}
        staticProps={props}
      />
    </>
  )
}

export async function getStaticPaths() {
  const paths = await generateStaticPaths({
    localContentDir,
    navDataFile,
    remotePluginsFile,
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
    remotePluginsFile,
  })
  return { props }
}

export default DocsLayout
