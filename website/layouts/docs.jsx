import DocsPage from '@hashicorp/react-docs-page'
import order from 'data/docs-navigation.js'
import { frontMatter as data } from '../pages/docs/**/*.mdx'
import Head from 'next/head'
import Link from 'next/link'
import { createMdxProvider } from '@hashicorp/nextjs-scripts/lib/providers/docs'

const MDXProvider = createMdxProvider({ product: 'packer' })

export default function DocsLayoutWrapper(pageMeta) {
  function DocsLayout(props) {
    return (
      <MDXProvider>
        <DocsPage
          {...props}
          product="packer"
          head={{
            is: Head,
            title: `${pageMeta.page_title} | Packer by HashiCorp`,
            description: pageMeta.description,
            siteName: 'Packer by HashiCorp',
          }}
          sidenav={{
            Link,
            category: 'docs',
            currentPage: props.path,
            data,
            order,
          }}
          resourceURL={`https://github.com/hashicorp/packer/blob/master/website/pages/${pageMeta.__resourcePath}`}
        />
      </MDXProvider>
    )
  }

  DocsLayout.getInitialProps = ({ asPath }) => ({ path: asPath })

  return DocsLayout
}
