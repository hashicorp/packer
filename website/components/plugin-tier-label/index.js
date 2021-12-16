import React from 'react'
import InlineSvg from '@hashicorp/react-inline-svg'
import svgRibbonIcon from './ribbon-icon.svg?include'
import classnames from 'classnames'
import s from './style.module.css'

const tierNames = {
  official: 'Official',
  community: 'Community',
}

function PluginTierLabel({ tier, isPageHeading = false }) {
  return (
    <div
      className={classnames(s.root, s[`tier-${tier}`], {
        [s.isPageHeading]: isPageHeading,
      })}
    >
      {tier === 'official' ? (
        <InlineSvg className={s.icon} src={svgRibbonIcon} />
      ) : null}
      <span className={s.text}>{tierNames[tier]}</span>
    </div>
  )
}

export default PluginTierLabel
