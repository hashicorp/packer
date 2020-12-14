import VERSION, { packageManagers } from 'data/version.js'
import Head from 'next/head'
import HashiHead from '@hashicorp/react-head'
import ProductDownloader from '@hashicorp/react-product-downloader'
import styles from './style.module.css'
import logo from '@hashicorp/mktg-assets/dist/product/packer-logo/color.svg'

export default function DownloadsPage({ releases }) {
  return (
    <>
      <HashiHead is={Head} title={`Downloads | Packer by HashiCorp`} />

      <ProductDownloader
        releases={releases}
        packageManagers={packageManagers}
        productName="Packer"
        productId="packer"
        latestVersion={VERSION}
        getStartedDescription="Follow step-by-step tutorials on AWS, Azure, GCP, and localhost."
        getStartedLinks={[
          {
            label: 'Placeholder',
            href: '#',
          },
          {
            label: 'Placeholder',
            href: '#',
          },
        ]}
        logo={<img className={styles.logo} alt="Packer" src={logo} />}
        brand="packer"
        tutorialLink={{
          href: 'https://learn.hashicorp.com/packer',
          label: 'View Tutorials at HashiCorp Learn',
        }}
      />
    </>
  )
}

export async function getStaticProps() {
  return fetch(`https://releases.hashicorp.com/packer/index.json`, {
    headers: {
      'Cache-Control': 'no-cache',
    },
  })
    .then((res) => res.json())
    .then((result) => {
      return {
        props: {
          releases: result,
        },
      }
    })
    .catch(() => {
      throw new Error(
        `--------------------------------------------------------
        Unable to resolve version ${VERSION} on releases.hashicorp.com from link
        <https://releases.hashicorp.com/packer/${VERSION}/index.json>. Usually this
        means that the specified version has not yet been released. The downloads page
        version can only be updated after the new version has been released, to ensure
        that it works for all users.
        ----------------------------------------------------------`
      )
    })
}
