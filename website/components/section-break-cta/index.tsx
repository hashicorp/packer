import Image from 'next/image'
import Button from '@hashicorp/react-button'
import InlineSvg from '@hashicorp/react-inline-svg'
import s from './style.module.css'

export default function SectionBreakCta({
  badge,
  heading,
  description,
  link,
  media,
}) {
  return (
    <div className={s.sectionBreakCta}>
      <div className={s.content}>
        <div className={s.eyebrow}>
          <InlineSvg
            className={s.logo}
            src={require('./hcp-packer.svg?include')}
          />
          <span className={s.badge}>{badge}</span>
        </div>
        <h2 className={s.heading}>{heading}</h2>
        <p className={s.description}>{description}</p>
        <Button
          title={link.text}
          url={link.url}
          theme={{
            brand: 'packer',
          }}
        />
      </div>
      <div className={s.media}>
        <Image
          src={media.src}
          width={media.width}
          height={media.height}
          alt={media.alt}
          layout="responsive"
        />
      </div>
    </div>
  )
}
