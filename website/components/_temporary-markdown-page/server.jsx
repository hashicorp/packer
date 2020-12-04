import fs from 'fs'
import path from 'path'
import renderToString from 'next-mdx-remote/render-to-string'
import markdownDefaults from '@hashicorp/nextjs-scripts/markdown'
import matter from 'gray-matter'

export default function generateGetStaticProps({
  pagePath,
  includesRoot = path.join(process.cwd(), 'content/partials'),
}) {
  return async function getStaticProps() {
    const filePath = path.join(process.cwd(), pagePath)
    const fileContent = fs.readFileSync(filePath, 'utf8')
    const { data, content } = matter(fileContent)
    const mdxSource = await renderToString(content, {
      mdxOptions: markdownDefaults({ resolveIncludes: includesRoot }),
    })
    return {
      props: {
        staticProps: {
          mdxSource,
          head: { title: data.page_title, description: data.description },
        },
      },
    }
  }
}
