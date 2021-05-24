import s from './style.module.css'
import Button from '@hashicorp/react-button'

export default function BrandedCta({ heading, content, links }) {
  return (
    <div className={s.brandedCta}>
      <div className={`g-grid-container ${s.contentContainer}`}>
        <h2 className={`g-type-display-2 ${s.heading}`}>{heading}</h2>
        <div className="content-and-links">
          <p className={`g-type-body-large ${s.content}`}>{content}</p>
          <div className={s.links}>
            {links.map((link, stableIdx) => {
              return (
                <Button
                  // eslint-disable-next-line react/no-array-index-key
                  key={stableIdx}
                  linkType={link.type || ''}
                  theme={{
                    variant: stableIdx === 0 ? 'primary' : 'secondary',
                    brand: 'packer',
                    background: 'light',
                  }}
                  title={link.text}
                  url={link.url}
                />
              )
            })}
          </div>
        </div>
      </div>
    </div>
  )
}
