import React from 'react'
import InlineSvg from '@hashicorp/react-inline-svg'
import svgRibbonIcon from './ribbon-icon.svg?include'
import s from './style.module.css'

const tierNames = {
  official: 'Official',
  community: 'Community',
}

function PluginTierLabel({ tier }) {
  return (
    <div className={s.root} data-tier={tier}>
      {tier === 'official' ? (
        <InlineSvg className={s.icon} src={svgRibbonIcon} />
      ) : null}
      <span className={s.text}>{tierNames[tier]}</span>
    </div>
  )
}

export default PluginTierLabel
