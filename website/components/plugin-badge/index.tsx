import React from 'react'
import Badge, { BadgeTheme } from '../badge'
import svgRibbonIcon from './ribbon-icon.svg?include'
import svgCheckIcon from './check-icon.svg?include'

type PluginLabelType = 'official' | 'community' | 'hcp_packer_ready'

const badgeTypes = {
  official: {
    label: 'Official',
    theme: 'gold',
    iconSvg: svgRibbonIcon,
  },
  community: {
    label: 'Community',
    theme: 'gray',
    iconSvg: false,
  },
  hcp_packer_ready: {
    label: 'HCP Packer Ready',
    theme: 'blue',
    iconSvg: svgCheckIcon,
  },
}

function PluginBadge({ type }: { type: PluginLabelType }): React.ReactElement {
  const { label, theme, iconSvg } = badgeTypes[type]
  return <Badge label={label} theme={theme as BadgeTheme} iconSvg={iconSvg} />
}

export default PluginBadge
