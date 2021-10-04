import Alert from '@hashicorp/react-alert'
import Button from '@hashicorp/react-button'
import s from './style.module.css'

export default function HomepageHero({
  heading,
  heroFeature,
  subheading,
  links,
  alert,
}) {
  return (
    <div className={s.homepageHero}>
      <div className={s.gridContainer}>
        <div className={s.content}>
          {alert ? (
            // @ts-expect-error -- prop types are incorrect, state is not needed
            <Alert
              url={alert.url}
              tag={alert.tag}
              product="packer"
              text={alert.text}
              textColor="dark"
            />
          ) : null}
          <h1 className={s.heading}>{heading}</h1>
          <p className={s.subheading}>{subheading}</p>
          <div className={s.links}>
            {links.map((link, index) => (
              <Button
                key={link.text}
                title={link.text}
                linkType={link.type}
                url={link.url}
                theme={{
                  variant: index === 0 ? 'primary' : 'secondary',
                  brand: index === 0 ? 'packer' : 'neutral',
                }}
              />
            ))}
          </div>
        </div>
        <div className={s.heroFeature}>
          <div className={s.heroFeatureFrame}>{heroFeature}</div>
        </div>
      </div>
    </div>
  )
}
