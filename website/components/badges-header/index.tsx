import PluginTierLabel, { PluginTierType } from '../plugin-tier-label'
import s from './style.module.css'

function BadgesHeader({
  badges,
}: {
  badges: PluginTierType[]
}): React.ReactElement {
  return (
    <div className={s.root}>
      <div className={s.surroundSpaceCompensator}>
        {badges.map((tierSlug, idx) => {
          return (
            <div className={s.badgeSpacer} key={idx}>
              <PluginTierLabel tier={tierSlug} />
            </div>
          )
        })}
      </div>
    </div>
  )
}

export default BadgesHeader
