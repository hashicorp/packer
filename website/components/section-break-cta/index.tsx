import Button from '@hashicorp/react-button'
import s from './style.module.css'

export default function SectionBreakCta({ heading, link }) {
  return (
    <div className={s.sectionBreakCta}>
      <hr />
      <h4 className={s.heading}>{heading}</h4>
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
