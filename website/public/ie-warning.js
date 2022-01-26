!(function () {
  'use strict'

  function applyStyles(e, s) {
    e.style.cssText = s
      .map(function (r) {
        return r.join(':')
      })
      .join(';')
  }

  const el = document.createElement('div')
  const containerStyles = [
    ['background-color', '#FCF0F2'],
    ['border-bottom', '1px solid #FFD4D6'],
    ['color', '#BA2226'],
    ['text-align', 'center'],
    ['font-family', '"Segoe UI", sans-serif'],
    ['font-weight', 'bold'],
  ]
  applyStyles(el, containerStyles)

  const message = document.createElement('p')
  const textStyles = [
    ['padding', '16px 0'],
    ['margin', '0'],
    ['color', '#BA2226'],
  ]
  applyStyles(message, textStyles)
  message.textContent = 'Internet Explorer is no longer supported. '

  const link = document.createElement('a')
  link.textContent = 'Learn more.'
  link.href = 'https://www.microsoft.com/en-us/edge?form=MA13DL&OCID=MA13DL'

  message.appendChild(link)
  el.appendChild(message)

  document.body.insertBefore(el, document.body.childNodes[0])
})()
