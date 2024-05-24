/**
 * Define your custom redirects within this file.
 *
 * Vercel's redirect documentation:
 * https://nextjs.org/docs/api-reference/next.config.js/redirects
 *
 * Relative paths with fragments (#) are not supported.
 * For destinations with fragments, use an absolute URL.
 *
 * Playground for testing url pattern matching: https://npm.runkit.com/path-to-regexp
 *
 * Note that redirects defined in a product's redirects file are applied to
 * the developer.hashicorp.com domain, which is where the documentation content
 * is rendered. Redirect sources should be prefixed with the product slug
 * to ensure they are scoped to the product's section. Any redirects that are
 * not prefixed with a product slug will be ignored.
 */
module.exports = [
  /*
  Example redirect:
  {
    source: '/packer/docs/internal-docs/my-page',
    destination: '/packer/docs/internals/my-page',
    permanent: true,
  },
  */
  /**
   * BEGIN EMPTY PAGE REDIRECTS
   * These redirects ensure some empty placeholder pages, dating back to when
   * "Overview" pages were a requirement, cannot be visited.
   *
   * These redirects can likely be removed once we have content API "pruning"
   * in place. That is, assuming the page at https://developer.hashicorp.com/packer/docs/templates/hcl_templates/functions/conversion
   * is still empty, the content API response from the content URL for that page
   * (https://content.hashicorp.com/api/content/packer/doc/latest/docs/templates/hcl_templates/functions/conversion)
   * should be a 404. Asana task for this "don't return content for empty" work:
   * https://app.asana.com/0/1100423001970639/1202110665886351/f
   */
  {
    source: '/packer/docs/templates/hcl_templates/functions/collection',
    destination: '/packer/docs/templates/hcl_templates/functions',
    permanent: true,
  },
  {
    source: '/packer/docs/templates/hcl_templates/functions/contextual',
    destination: '/packer/docs/templates/hcl_templates/functions',
    permanent: true,
  },
  {
    source: '/packer/docs/templates/hcl_templates/functions/conversion',
    destination: '/packer/docs/templates/hcl_templates/functions',
    permanent: true,
  },
  {
    source: '/packer/docs/templates/hcl_templates/functions/crypto',
    destination: '/packer/docs/templates/hcl_templates/functions',
    permanent: true,
  },
  {
    source: '/packer/docs/templates/hcl_templates/functions/encoding',
    destination: '/packer/docs/templates/hcl_templates/functions',
    permanent: true,
  },
  {
    source: '/packer/docs/templates/hcl_templates/functions/file',
    destination: '/packer/docs/templates/hcl_templates/functions',
    permanent: true,
  },
  {
    source: '/packer/docs/templates/hcl_templates/functions/ipnet',
    destination: '/packer/docs/templates/hcl_templates/functions',
    permanent: true,
  },
  {
    source: '/packer/docs/templates/hcl_templates/functions/numeric',
    destination: '/packer/docs/templates/hcl_templates/functions',
    permanent: true,
  },
  {
    source: '/packer/docs/templates/hcl_templates/functions/string',
    destination: '/packer/docs/templates/hcl_templates/functions',
    permanent: true,
  },
  {
    source: '/packer/docs/templates/hcl_templates/functions/uuid',
    destination: '/packer/docs/templates/hcl_templates/functions',
    permanent: true,
  },
  {
    source: '/packer/docs/plugins/install-plugins',
    destination: '/packer/docs/plugins/install',
    permanent: true,
  },
  /**
   * END EMPTY PAGE REDIRECTS
   */
]
