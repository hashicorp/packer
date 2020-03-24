// The root folder for this documentation category is `pages/intro`
//
// - A string refers to the name of a file
// - A "category" value refers to the name of a directory
// - All directories must have an "index.mdx" file to serve as
//   the landing page for the category

export default [
  'why',
  'use-cases',
  {
    category: 'getting-started',
    content: ['build-image', 'provision', 'parallel-builds', 'vagrant', 'next']
  }
]
