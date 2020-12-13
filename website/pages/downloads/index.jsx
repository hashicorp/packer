import VERSION from 'data/version.js'
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
        packageManagers={[
          {
            label: 'Homebrew',
            commands: [
              'brew tap hashicorp/tap',
              'brew install hashicorp/tap/packer',
            ],
            os: 'darwin',
          },
          {
            label: 'Ubuntu/Debian',
            commands: [
              'curl -fsSL https://apt.releases.hashicorp.com/gpg | sudo apt-key add -',
              'sudo apt-add-repository "deb [arch=amd64] https://apt.releases.hashicorp.com $(lsb_release -cs) main"',
              'sudo apt-get update && sudo apt-get install packer',
            ],
            os: 'linux',
          },
          {
            label: 'CentOS/RHEL',
            commands: [
              'sudo yum install -y yum-utils',
              'sudo yum-config-manager --add-repo https://rpm.releases.hashicorp.com/RHEL/hashicorp.repo',
              'sudo yum -y install packer',
            ],
            os: 'linux',
          },
          {
            label: 'Fedora',
            commands: [
              'sudo dnf install -y dnf-plugins-core',
              'sudo dnf config-manager --add-repo https://rpm.releases.hashicorp.com/fedora/hashicorp.repo',
              'sudo dnf -y install packer',
            ],
            os: 'linux',
          },
          {
            label: 'Amazon Linux',
            commands: [
              'sudo yum install -y yum-utils',
              'sudo yum-config-manager --add-repo https://rpm.releases.hashicorp.com/AmazonLinux/hashicorp.repo',
              'sudo yum -y install packer',
            ],
            os: 'linux',
          },
        ]}
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
