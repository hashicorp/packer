import ProductFeaturesList from '@hashicorp/react-product-features-list'
import AnimatedTerminal from 'components/animated-terminal'
import BrandedCta from 'components/branded-cta'
import HomepageHero from 'components/homepage-hero'
import IntegrationsTextSplit from 'components/integrations-text-split'
import SectionBreakCta from 'components/section-break-cta'

import s from './style.module.css'

export default function Homepage() {
  return (
    <div id="p-home" className={s.home}>
      <section id="hero">
        <HomepageHero
          alert={{
            url:
              'https://hashicorp.com/blog/hcp-packer-is-now-generally-available',
            tag: 'BLOG POST',
            text: 'Announcing HCP Packer Standard tier',
          }}
          heading="Build automated machine images"
          subheading="Create identical machine images for multiple platforms from a single source configuration."
          links={[
            {
              text: 'Try HCP Packer',
              url:
                'https://cloud.hashicorp.com/products/packer?utm_source=packer_io&utm_content=homepage_hero_packer',
            },
            {
              text: 'Install Packer',
              url: '/downloads',
            },
          ]}
          heroFeature={
            <AnimatedTerminal
              lines={[
                {
                  code: '$ packer build template.pkr.hcl',
                  color: 'gray',
                },
                {
                  code:
                    '==> virtualbox: virtualbox output will be in this color.',
                  color: 'white',
                },
                { code: '==> vmware: vmware output will be in this color.' },
                {
                  code:
                    '==> vmware: Copying or downloading ISO. Progress will be reported periodically.',
                },
                { code: '==> vmware: Creating virtual machine disk' },
                { code: '==> vmware: Building and writing VMX file' },
                { code: '==> vmware: Starting HTTP server on port 8964' },
                { code: '==> vmware: Starting virtual machine...' },
                {
                  code:
                    '==> virtualbox: Downloading VirtualBox guest additions. Progress will be shown periodically',
                  color: 'white',
                },
                {
                  code:
                    '==> virtualbox: Copying or downloading ISO. Progress will be reported periodically.',
                  color: 'white',
                },
                {
                  code: '==> virtualbox: Starting HTTP server on port 8081',
                  color: 'white',
                },
                {
                  code: '==> virtualbox: Creating virtual machine...',
                  color: 'white',
                },
                {
                  code: '==> virtualbox: Creating hard drive...',
                  color: 'white',
                },
              ]}
            />
          }
        />
      </section>
      <section id="features">
        <ProductFeaturesList
          heading="Why Packer?"
          features={[
            {
              title: 'Rapid Infrastructure Deployment',
              content:
                'Use Terraform to launch completely provisioned and configured machine instances with Packer images in seconds.',
              icon: '/img/product-features-list/deployment.svg',
            },
            {
              title: 'Multi-provider Portability',
              content:
                'Identical images allow you to run dev, staging, and production environments across platforms.',
              icon: '/img/product-features-list/portability.svg',
            },
            {
              title: 'Improved Stability',
              content:
                'By provisioning instances from stable images installed and configured by Packer, you can ensure buggy software does not get deployed.',
              icon: '/img/product-features-list/stability.svg',
            },
            {
              title: 'Increased Dev / Production Parity',
              content:
                'Keep dev, staging, and production environments as similar as possible by generating images for multiple platforms at the same time.',
              icon: '/img/product-features-list/prod-parity.svg',
            },
            {
              title: 'Reliable Continuous Delivery',
              content:
                'Generate new machine images for multiple platforms, launch and test, and verify the infrastructure changes work; then, use Terraform to put your images in production.',
              icon: '/img/product-features-list/continuous-delivery.svg',
            },
            {
              title: 'Appliance Demo Creation',
              content:
                'Create software appliances and disposable product demos quickly, even with software that changes continuously.',
              icon: '/img/product-features-list/demo-creation.svg',
            },
          ]}
        />
      </section>
      <section className={s.sectionGridContainer}>
        <SectionBreakCta
          badge="Beta"
          heading="Automate build configuration across cloud providers."
          description="Set up HCP Packer in minutes to start tracking Packer images across your provisioning pipeline."
          link={{
            text: 'Learn more',
            url:
              'https://cloud.hashicorp.com/products/packer?utm_source=packer_io&utm_content=homepage_signup_callout',
          }}
          media={{
            src: '/img/hcp-callout.svg',
            width: 646,
            height: 441,
            alt: 'Packer UI',
          }}
        />
      </section>
      <section>
        <IntegrationsTextSplit
          heading="Extending Packer with Plugins"
          content={
            <>
              <p className="g-type-body">
                Extend Packer’s functionality without modifying Packer core.
                Plugins are capable of adding these components:
              </p>
              <ul className={s.textSplitList}>
                <li>Builders</li>
                <li>Provisioners</li>
                <li>Post-processors</li>
                <li>Data sources</li>
              </ul>
            </>
          }
          links={[
            {
              text: 'Read the Docs',
              url: '/docs',
              type: 'none',
            },
            {
              text: 'Develop a plugin',

              url: '/docs/plugins',
              type: 'inbound',
            },
          ]}
          image={{
            url: '/img/integrations-text-split/integrations.png',
            alt: 'Build images with Packer plugins',
          }}
        />
      </section>
      <section id="get-started">
        <BrandedCta
          heading="Ready to get started?"
          content="Start by following a tutorial to create a simple vm image with Packer or learn about how the project works by exploring the documentation."
          links={[
            {
              text: 'Get Started',
              url: 'https://learn.hashicorp.com/packer',
              type: 'outbound',
            },
            { text: 'Explore documentation', url: '/docs' },
          ]}
        />
      </section>
    </div>
  )
}
