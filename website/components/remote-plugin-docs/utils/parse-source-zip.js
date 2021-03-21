const path = require('path')
const AdmZip = require('adm-zip')
const validatePluginDocsFiles = require('./validate-plugin-docs-files')

/*

NOTE: used for fallback approach, where we parse from
the full release archive

*/

// Given a response from fetching a source .zip file,
// which contains a "docs" folder,
//
// return [null, docsMdxFiles] if docs files
// are successfully fetched and valid,
// where docsMdxFiles is an array of { filePath, fileString } items.
//
// otherwise, return [err, null]
// where err is an error message describing whether the
// docs files were missing or invalid, with a path to resolution
async function parseSourceZip(response) {
  const responseBuffer = Buffer.from(await response.arrayBuffer())
  const responseZip = new AdmZip(responseBuffer)
  const sourceEntries = responseZip.getEntries()
  const docsEntries = sourceEntries.filter((entry) => {
    // filter for zip entries in the docs subfolder only
    const dirs = path.dirname(entry.entryName).split('/')
    return dirs.length > 1 && dirs[1] === 'docs'
  })
  // Validate the file paths within the "docs" folder
  const docsFilePaths = docsEntries.map((e) => {
    // We need to remove the leading directory,
    // which will be something like packer-plugin-docs-0.0.5
    const filePath = e.entryName.split('/').slice(1).join('/')
    return filePath
  })
  const validationError = validatePluginDocsFiles(docsFilePaths)
  if (validationError) return [validationError, null]
  // If valid, filter for MDX files only, and return
  // a { filePath, fileString } object for each mdx file
  const docsMdxFiles = docsEntries
    .filter((e) => {
      return path.extname(e.entryName) === '.mdx'
    })
    .map((e) => {
      // We need to remove the leading directory,
      // which will be something like packer-plugin-docs-0.0.5
      const filePath = e.entryName.split('/').slice(1).join('/')
      const fileString = e.getData().toString()
      return { filePath, fileString }
    })
  return [null, docsMdxFiles]
}

module.exports = parseSourceZip
