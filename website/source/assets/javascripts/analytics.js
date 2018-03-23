document.addEventListener('DOMContentLoaded', function() {
  track('.downloads .download a', function(el) {
    return {
      event: 'Download',
      category: 'Button',
      label: 'Packer | v' + el.href.match(/\/(\d+\.\d+\.\d+)\//)[1]
    }
  })
})
