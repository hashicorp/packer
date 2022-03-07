import React from 'react'
import InlineSvg from '@hashicorp/react-inline-svg'
import classnames from 'classnames'
import s from './style.module.css'

type BadgeTheme = 'gray' | 'blue' | 'gold' | 'purple' | 'light-gray'

interface BadgeProps {
  label: string
  iconSvg?: string
  theme?: BadgeTheme
  href?: string
}

function Badge({
  theme = 'gray',
  label,
  iconSvg,
  href,
}: BadgeProps): React.ReactElement {
  const Elem = href ? 'a' : 'div'
  return (
    <Elem href={href} className={classnames(s.root, s[`theme-${theme}`])}>
      {iconSvg ? <InlineSvg className={s.icon} src={iconSvg} /> : null}
      <span className={s.text}>{label}</span>
    </Elem>
  )
}

export type { BadgeTheme }
export default Badge
