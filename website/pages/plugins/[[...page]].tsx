import { productName, productSlug } from 'data/metadata'
import DocsPage from '@hashicorp/react-docs-page'
import { NextPage, InferGetStaticPropsType } from 'next'

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
const baseRoute = 'plugins'
const localContentDir = 'content/plugins'
const mainBranch = 'master'
const navDataFile = 'data/plugins-nav-data.json'
const product = { name: productName, slug: productSlug }
const remotePluginsFile = 'data/plugins-manifest.json'

type Props = InferGetStaticPropsType<typeof getStaticProps>
const DocsLayout: NextPage<Props> = ({
  isDevMissingRemotePlugins,
  ...props
}) => {
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
        // @ts-expect-error
        staticProps={props}
        showVersionSelect={false}
      />
    </>
  )
}

export async function getStaticPaths() {
  const paths = await generateStaticPaths({
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
