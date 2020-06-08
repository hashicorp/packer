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
        title: 'Overview',
        href: 'https://learn.hashicorp.com/packer/getting-started/install',
      },
      {
        title: 'Build An Image',
        href: 'https://learn.hashicorp.com/packer/getting-started/build-image',
      },
      {
        title: 'Provision',
        href: 'https://learn.hashicorp.com/packer/getting-started/provision',
      },
      {
        title: 'Parallel Builds',
        href:
          'https://learn.hashicorp.com/packer/getting-started/parallel-builds',
      },
      {
        title: 'Vagrant Boxes',
        href: 'https://learn.hashicorp.com/packer/getting-started/vagrant',
      },
      {
        title: 'Next Steps',
        href: 'https://learn.hashicorp.com/packer/getting-started/next',
      },
    ],
  },
]
