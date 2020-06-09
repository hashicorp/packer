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
    name: 'Getting Started',
    content: [
      {
        title: 'Install',
        href: '/intro/getting-started/install',
      },
      {
        title: 'Build An Image',
        href: '/intro/getting-started/build-image',
      },
      {
        title: 'Provision',
        href: '/intro/getting-started/provision',
      },
      {
        title: 'Parallel Builds',
        href: '/intro/getting-started/parallel-builds',
      },
      {
        title: 'Vagrant Boxes',
        href: '/intro/getting-started/vagrant',
      },
      {
        title: 'Next Steps',
        href: '/intro/getting-started/next',
      },
    ],
  },
]
