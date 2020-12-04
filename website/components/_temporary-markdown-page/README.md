# `<MarkdownPage />`

This component renders a single page built from a separate markdown file.

## Usage

```jsx
import MarkdownPage from '@hashicorp/react-markdown-page'
import generateStaticProps from '@hashicorp/react-markdown-page/server'

export default function MyPage({ staticProps }) {
  return <MarkdownPage {...staticProps} />
}

export function getStaticProps() {
  return generateStaticProps({
    pagePath: 'content/test-page.mdx', // resolved from project root
  })
}
```

If the specified page contains front matter, the `page_title` and `description` keys will be added as the title and description the the `<head>` of the page.
