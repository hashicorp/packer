import s from './style.module.css'

function DevAlert({ children }) {
  return (
    <div className={s.root}>
      <div className={s.inner}>{children}</div>
    </div>
  )
}

export default DevAlert
