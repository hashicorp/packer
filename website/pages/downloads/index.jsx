import VERSION from 'data/version'
import { productSlug } from 'data/metadata'
import ProductDownloadsPage from '@hashicorp/react-product-downloads-page'
import { generateStaticProps } from '@hashicorp/react-product-downloads-page/server'
import s from './style.module.css'

export default function DownloadsPage(staticProps) {
  return (
    <ProductDownloadsPage
      getStartedDescription="Follow step-by-step tutorials on the essentials of Packer."
      getStartedLinks={[
        {
          label: 'View all Packer tutorials',
          href: 'https://learn.hashicorp.com/packer',
        },
      ]}
      logo={
        <img
          className={s.logo}
          alt="Packer"
          src={require('./img/packer-logo.svg')}
        />
      }
      tutorialLink={{
        href: 'https://learn.hashicorp.com/packer',
        label: 'View Tutorials at HashiCorp Learn',
      }}
      {...staticProps}
    />
  )
}

export async function getStaticProps() {
  return generateStaticProps({
    product: productSlug,
    latestVersion: VERSION,
  })
}
