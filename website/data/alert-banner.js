export const ALERT_BANNER_ACTIVE = true

// https://github.com/hashicorp/web-components/tree/master/packages/alert-banner
export default {
  tag: 'New',
  url: 'https://cloud.hashicorp.com/products/packer ',
  text:
    'HCP Packer offers automation and security workflows for Packer, and is now generally available.',
  linkText: 'Sign up for free',
  // Set the expirationDate prop with a datetime string (e.g. '2020-01-31T12:00:00-07:00')
  // if you'd like the component to stop showing at or after a certain date
  expirationDate: '2022-04-07T23:00:00-07:00',
}
