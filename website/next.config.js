const withHashicorp = require('@hashicorp/nextjs-scripts')
const path = require('path')
const redirects = require('./redirects.next')

module.exports = withHashicorp({
  defaultLayout: true,
  transpileModules: ['is-absolute-url', '@hashicorp/react-.*'],
  mdx: { resolveIncludes: path.join(__dirname, 'pages/partials') },
})({
  svgo: { plugins: [{ removeViewBox: false }] },
  rewrites: () => [
    {
      source: '/api/:path*',
      destination: '/api-docs/:path*',
    },
  ],
  redirects: () => redirects,
  // Note: These are meant to be public, it's not a mistake that they are here
  env: {
    HASHI_ENV: process.env.HASHI_ENV,
    SEGMENT_WRITE_KEY: 'AjXdfmTTk1I9q9dfyePuDFHBrz1tCO3l',
    BUGSNAG_CLIENT_KEY: 'de0b822b269aa57b620efd8927e03744',
    BUGSNAG_SERVER_KEY: 'b6c57b27a37e531a5de94f065dd98bc0',
  },
})
