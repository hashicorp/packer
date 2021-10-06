import './style.css'
import '@hashicorp/platform-util/nprogress/style.css'

import ProductSubnav from 'components/subnav'
import HashiStackMenu from '@hashicorp/react-hashi-stack-menu'
import Footer from 'components/footer'
import Error from './_error'
import Head from 'next/head'
import HashiHead from '@hashicorp/react-head'
import { useEffect } from 'react'
import Router, { useRouter } from 'next/router'
import NProgress from '@hashicorp/platform-util/nprogress'
import createConsentManager from '@hashicorp/react-consent-manager/loader'
import { ErrorBoundary } from '@hashicorp/platform-runtime-error-monitoring'
import useAnchorLinkAnalytics from '@hashicorp/platform-util/anchor-link-analytics'
import AlertBanner from '@hashicorp/react-alert-banner'
import alertBannerData, { ALERT_BANNER_ACTIVE } from 'data/alert-banner'

NProgress({ Router })
const { ConsentManager, openConsentManager } = createConsentManager({
  preset: 'oss',
})

export default function App({ Component, pageProps }) {
  const router = useRouter()

  useEffect(() => {
    // Load Fathom analytics
    Fathom.load('WYIVNEGX', {
      includedDomains: ['packer.io', 'www.packer.io'],
    })

    function onRouteChangeComplete() {
      Fathom.trackPageview()
    }

    // Record a pageview when route changes
    router.events.on('routeChangeComplete', onRouteChangeComplete)

    // Unassign event listener
    return () => {
      router.events.off('routeChangeComplete', onRouteChangeComplete)
    }
  }, [])

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
        icon={[{ href: '/favicon.ico' }]}
      />
      {ALERT_BANNER_ACTIVE && (
        <AlertBanner {...alertBannerData} product="packer" hideOnMobile />
      )}
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
