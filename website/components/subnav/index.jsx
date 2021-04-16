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
        { text: 'Download', url: '/downloads' },
      ]}
      currentPath={router.asPath}
      menuItemsAlign="right"
      menuItems={subnavItems}
      constrainWidth
    />
  )
}
