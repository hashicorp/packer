import './style.css'
import '@hashicorp/nextjs-scripts/lib/nprogress/style.css'

import ProductSubnav from 'components/subnav'
import HashiStackMenu from '@hashicorp/react-hashi-stack-menu'
import Footer from 'components/footer'
import Error from './_error'
import Head from 'next/head'
import HashiHead from '@hashicorp/react-head'
import Router from 'next/router'
import NProgress from '@hashicorp/nextjs-scripts/lib/nprogress'
import createConsentManager from '@hashicorp/nextjs-scripts/lib/consent-manager'
import { ErrorBoundary } from '@hashicorp/nextjs-scripts/lib/bugsnag'
import useAnchorLinkAnalytics from '@hashicorp/nextjs-scripts/lib/anchor-link-analytics'

NProgress({ Router })
const { ConsentManager, openConsentManager } = createConsentManager({
  preset: 'oss',
})

export default function App({ Component, pageProps }) {
  useAnchorLinkAnalytics()

  return (
    <ErrorBoundary FallbackComponent={Error}>
      <HashiHead
        is={Head}
        title="Packer by HashiCorp"
        siteName="Packer by HashiCorp"
        description="Packer is a free and open source tool for creating golden images for multiple
        platforms from a single source configuration."
        image="https://www.packer.io/img/og-image.png"
        icon={[{ href: '/_favicon.ico' }]}
      />
      <HashiStackMenu />
      <ProductSubnav />
      <div className="content">
        <Component {...pageProps} />
      </div>
      <Footer openConsentManager={openConsentManager} />
      <ConsentManager />
    </ErrorBoundary>
  )
}
