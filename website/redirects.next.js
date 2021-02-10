module.exports = [
  {
    source: '/docs/installation',
    destination: '/docs/install',
    permanent: true,
  },
  {
    source: '/docs/command-line/machine-readable',
    destination: '/docs/commands',
    permanent: true,
  },
  {
    source: '/docs/command-line/introduction',
    destination: '/docs/commands',
    permanent: true,
  },
  {
    source: '/docs/templates/introduction',
    destination: '/docs/templates',
    permanent: true,
  },
  {
    source: '/docs/builders/azure-setup',
    destination: '/docs/builders/azure',
    permanent: true,
  },
  {
    source: '/docs/templates/veewee-to-packer',
    destination: '/guides/veewee-to-packer',
    permanent: true,
  },
  {
    source: '/docs/extend/developing-plugins',
    destination: '/docs/plugins',
    permanent: true,
  },
  {
    source: '/docs/extending/developing-plugins',
    destination: '/docs/plugins',
    permanent: true,
  },
  {
    source: '/docs/extending/plugins',
    destination: '/docs/plugins',
    permanent: true,
  },
  {
    source: '/docs/extend/builder',
    destination: '/docs/extending/custom-builders',
    permanent: true,
  },
  {
    source: '/docs/extending/custom-builders',
    destination: '/docs/plugins/creation/custom-builders',
    permanent: true,
  },
  {
    source: '/docs/extending/custom-provisioners',
    destination: '/docs/plugins/creation/custom-provisioners',
    permanent: true,
  },
  {
    source: '/docs/extending/custom-post-processors',
    destination: '/docs/plugins/creation/custom-post-processors',
    permanent: true,
  },
  {
    source: '/docs/getting-started/setup',
    destination: '/docs/getting-started/install',
    permanent: true,
  },
  {
    source: '/docs/other/community',
    destination: '/community-tools',
    permanent: true,
  },
  {
    source: '/downloads-community',
    destination: '/community-tools',
    permanent: true,
  },
  {
    source: '/docs/platforms',
    destination: '/docs/builders',
    permanent: true,
  },
  {
    source: '/intro/platforms',
    destination: '/docs/builders',
    permanent: true,
  },
  {
    source: '/docs/templates/configuration-templates',
    destination: '/docs/templates/legacy_json_templates/engine',
    permanent: true,
  },
  {
    source: '/docs/machine-readable/:path*',
    destination: '/docs/commands',
    permanent: true,
  },
  {
    source: '/docs/command-line/:path*',
    destination: '/docs/commands/:path*',
    permanent: true,
  },
  {
    source: '/intro/getting-started',
    destination: 'https://learn.hashicorp.com/packer/getting-started/install',
    permanent: true,
  },
  {
    source: '/intro/getting-started/install',
    destination: 'https://learn.hashicorp.com/packer/getting-started/install',
    permanent: true,
  },
  {
    source: '/intro/getting-started/build-image',
    destination:
      'https://learn.hashicorp.com/packer/getting-started/build-image',
    permanent: true,
  },
  {
    source: '/intro/getting-started/provision',
    destination: 'https://learn.hashicorp.com/packer/getting-started/provision',
    permanent: true,
  },
  {
    source: '/intro/getting-started/parallel-builds',
    destination:
      'https://learn.hashicorp.com/packer/getting-started/parallel-builds',
    permanent: true,
  },
  {
    source: '/intro/getting-started/vagrant',
    destination: 'https://learn.hashicorp.com/packer/getting-started/vagrant',
    permanent: true,
  },
  {
    source: '/intro/getting-started/next',
    destination: 'https://learn.hashicorp.com/packer/getting-started/next   ',
    permanent: true,
  },
  {
    source: '/docs/basics/terminology',
    destination: '/docs/terminology',
    permanent: true,
  },
  {
    source: '/docs/other/:path*',
    destination: '/docs/:path*',
    permanent: true,
  },
  {
    source: '/docs/configuration/from-1.5/:path*/overview',
    destination: '/docs/templates/hcl_templates/:path*',
    permanent: true,
  },
  {
    source: '/docs/from-1.5/:path*',
    destination: '/docs/templates/hcl_templates/:path*',
    permanent: true,
  },
  {
    source: '/docs/templates/hcl_templates/:path*/overview',
    destination: '/docs/templates/hcl_templates/:path*',
    permanent: true,
  },
  {
    source: '/docs/builders/amazon-:path',
    destination: '/docs/builders/amazon/:path*',
    permanent: true,
  },
  {
    source: '/docs/builders/azure-:path',
    destination: '/docs/builders/azure/:path*',
    permanent: true,
  },
  {
    source: '/docs/builders/hyperv-:path',
    destination: '/docs/builders/hyperv/:path*',
    permanent: true,
  },
  {
    source: '/docs/builders/oracle-:path',
    destination: '/docs/builders/oracle/:path*',
    permanent: true,
  },
  {
    source: '/docs/builders/osc-:path',
    destination: '/docs/builders/outscale/:path*',
    permanent: true,
  },
  {
    source: '/docs/builders/parallels-:path',
    destination: '/docs/builders/parallels/:path*',
    permanent: true,
  },
  {
    source: '/docs/builders/virtualbox-:path',
    destination: '/docs/builders/virtualbox/:path*',
    permanent: true,
  },
  {
    source: '/docs/builders/vmware-:path',
    destination: '/docs/builders/vmware/:path*',
    permanent: true,
  },
  {
    source: '/docs/builders/vsphere-:path',
    destination: '/docs/builders/vmware/vsphere-:path*',
    permanent: true,
  },
  // disallow '.html' or '/index.html' in favor of cleaner, simpler paths
  { source: '/:path*/index', destination: '/:path*', permanent: true },
  { source: '/:path*.html', destination: '/:path*', permanent: true },
]
