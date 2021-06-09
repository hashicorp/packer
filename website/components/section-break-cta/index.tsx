import Button from '@hashicorp/react-button'
import s from './style.module.css'

export default function SectionBreakCta({ heading, description, link }) {
  return (
    <div className={s.sectionBreakCta}>
      <hr />
      <h4 className={s.heading}>{heading}</h4>
      {description && <p className={s.description}>{description}</p>}
      <Button
        title={link.text}
        url={link.url}
        theme={{
          brand: 'neutral',
          variant: 'tertiary-neutral',
          background: 'light',
        }}
        linkType="outbound"
      />
    </div>
  )
}
