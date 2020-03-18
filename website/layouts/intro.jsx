import DocsPage from '@hashicorp/react-docs-page'
import order from '../data/intro-navigation.js'
import { frontMatter as data } from '../pages/intro/**/*.mdx'
import Head from 'next/head'
import Link from 'next/link'

function IntroLayoutWrapper(pageMeta) {
  function IntroLayout(props) {
    return (
      <DocsPage
        {...props}
        product="packer"
        head={{
          is: Head,
          title: `${pageMeta.page_title} | Packer by HashiCorp`,
          description: pageMeta.description,
          siteName: 'Packer by HashiCorp'
        }}
        sidenav={{
          Link,
          category: 'guides',
          currentPage: props.path,
          data,
          order
        }}
        resourceURL={`https://github.com/hashicorp/packer/blob/master/website/pages/${pageMeta.__resourcePath}`}
      />
    )
  }

  IntroLayout.getInitialProps = ({ asPath }) => ({ path: asPath })

  return IntroLayout
}

export default IntroLayoutWrapper
