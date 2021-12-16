import React from 'react'
import InlineSvg from '@hashicorp/react-inline-svg'
import svgRibbonIcon from './ribbon-icon.svg?include'
import svgCheckIcon from './check-icon.svg?include'
import classnames from 'classnames'
import s from './style.module.css'

type PluginTierType = 'official' | 'community' | 'hcp_packer_ready'

const tierNames = {
  official: 'Official',
  community: 'Community',
  hcp_packer_ready: 'HCP Packer Ready',
}

interface PluginTierLabelProps {
  tier: PluginTierType
  isPageHeading?: boolean
}

function PluginTierLabel({
  tier,
  isPageHeading = false,
}: PluginTierLabelProps): React.ReactElement {
  return (
    <div
      className={classnames(s.root, s[`tier-${tier}`], {
        [s.isPageHeading]: isPageHeading,
      })}
    >
      {tier === 'official' ? (
        <InlineSvg className={s.icon} src={svgRibbonIcon} />
      ) : tier === 'hcp_packer_ready' ? (
        <InlineSvg className={s.icon} src={svgCheckIcon} />
      ) : null}
      <span className={s.text}>{tierNames[tier]}</span>
    </div>
  )
}

export type { PluginTierType }
export default PluginTierLabel
