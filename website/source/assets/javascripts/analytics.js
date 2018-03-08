document.addEventListener('DOMContentLoaded', () => {
  track('.downloads .download a', el => {
    return {
      event: 'Download',
      category: 'Button',
      label: `Packer | v${el.href.match(/\/(\d+\.\d+\.\d+)\//)[1]}`
    }
  })
})
