// The root folder for this documentation category is `pages/guides`
//
// - A string refers to the name of a file
// - A "category" value refers to the name of a directory
// - All directories must have an "index.mdx" file to serve as
//   the landing page for the category

export default [
  {
    category: 'hcl',
    content: ['from-json-v1', 'variables', 'component-object-spec'],
  },
  {
    category: 'automatic-operating-system-installs',
    content: ['autounattend_windows', 'preseed_ubuntu'],
  },
  {
    category: '1.7-plugin-upgrade',
    content: [],
  },
  {
    category: 'workflow-tips-and-tricks',
    content: [
      'isotime-template-function',
      'veewee-to-packer',
      'use-packer-with-comment',
    ],
  },
  {
    category: 'packer-on-cicd',
    content: [
      'build-image-in-cicd',
      'build-virtualbox-image',
      'pipelineing-builds',
      'trigger-tfe',
      'upload-images-to-artifact',
    ],
  },
]
