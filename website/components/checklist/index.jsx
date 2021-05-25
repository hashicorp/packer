import s from './style.module.css'

export default function ChecklistWrapper({ children }) {
  return <div className={s.root}>{children}</div>
}
