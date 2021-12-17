import React from 'react'
import s from './style.module.css'

function BadgesHeader({
  children,
}: {
  children: React.ReactChild[]
}): React.ReactElement {
  const childrenArray = React.Children.toArray(children)
  return (
    <div className={s.root}>
      <div className={s.surroundSpaceCompensator}>
        {childrenArray.map((badge, idx) => {
          return (
            <div className={s.badgeSpacer} key={idx}>
              {badge}
            </div>
          )
        })}
      </div>
    </div>
  )
}

export default BadgesHeader
