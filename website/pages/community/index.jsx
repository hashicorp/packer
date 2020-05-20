import VerticalTextBlockList from '@hashicorp/react-vertical-text-block-list'
import SectionHeader from '@hashicorp/react-section-header'
import Head from 'next/head'

export default function CommunityPage() {
  return (
    <div id="p-community">
      <Head>
        <title key="title">Community | Packer by HashiCorp</title>
      </Head>
      <SectionHeader
        headline="Community"
        description="Packer is an open source project with a growing community. There are active, dedicated users willing to help you through various mediums."
        use_h1={true}
      />
      <VerticalTextBlockList
        data={[
          {
            header: 'Community Forum',
            body:
              '[Packer Community Forum](https://discuss.hashicorp.com/c/packer)',
          },
          {
            header: 'Announcement List',
            body:
              '[HashiCorp Announcement Google Group](https://groups.google.com/group/hashicorp-announce)',
          },
          {
            header: 'Bug Tracker',
            body:
              '[Issue tracker on GitHub](https://github.com/hashicorp/packer/issues). Please only use this for reporting bugs. For general help, please use the Community Forum.',
          },
          {
            header: 'Training',
            body:
              'Paid [HashiCorp training courses](https://www.hashicorp.com/training) are also available in a city near you. Private training courses are also available.',
          },
        ]}
      />
    </div>
  )
}
