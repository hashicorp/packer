import Subnav from '@hashicorp/react-subnav'
import subnavItems from 'data/subnav'
import { useRouter } from 'next/router'

export default function PackerSubnav() {
  const router = useRouter()
  return (
    <Subnav
      titleLink={{
        text: 'packer',
        url: '/',
      }}
      ctaLinks={[
        { text: 'GitHub', url: 'https://www.github.com/hashicorp/packer' },
        { text: 'Install Packer', url: '/downloads' },
        {
          text: 'Try HCP Packer',
          url:
            'https://cloud.hashicorp.com/products/packer?utm_source=packer_io&utm_content=top_nav_packer',
        },
      ]}
      hideGithubStars={true}
      currentPath={router.asPath}
      menuItemsAlign="right"
      menuItems={subnavItems}
      constrainWidth
      matchOnBasePath
    />
  )
}
