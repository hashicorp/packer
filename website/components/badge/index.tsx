import React from 'react'
import InlineSvg from '@hashicorp/react-inline-svg'
import classnames from 'classnames'
import s from './style.module.css'

type BadgeTheme = 'gray' | 'blue' | 'gold'

interface BadgeProps {
  label: string
  iconSvg?: string
  theme?: BadgeTheme
}

function Badge({
  theme = 'gray',
  label,
  iconSvg,
}: BadgeProps): React.ReactElement {
  return (
    <div className={classnames(s.root, s[`theme-${theme}`])}>
      {iconSvg ? <InlineSvg className={s.icon} src={iconSvg} /> : null}
      <span className={s.text}>{label}</span>
    </div>
  )
}

export type { BadgeTheme }
export default Badge
