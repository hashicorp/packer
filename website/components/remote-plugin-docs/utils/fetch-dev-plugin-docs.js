const path = require('path')
const validatePluginDocsFiles = require('./validate-plugin-docs-files')
const AdmZip = require('adm-zip')

// Given a zipFile path,
//
// return [null, docsMdxFiles] if docs files
// are successfully fetched and valid,
// where docsMdxFiles is an array of { filePath, fileString } items.
//
// otherwise, return [err, null]
// where err is an error message describing whether the
// docs files were missing or invalid, with a path to resolution
async function fetchDevPluginDocs(zipFile) {
  const [err, docsMdxFiles] = await parseZipFile(zipFile)
  if (err) {
    const errMsg = `Invalid plugin dev docs file ${zipFile}. ${err}`
    throw new Error(errMsg)
  }
  return docsMdxFiles
}

// Given a docs.zip filepath,
// which is a compressed "docs" folder,
//
// return [null, docsMdxFiles] if docs files
// are successfully fetched and valid,
// where docsMdxFiles is an array of { filePath, fileString } items.
//
// otherwise, return [err, null]
// where err is an error message describing whether the
// docs files were missing or invalid, with a path to resolution
async function parseZipFile(zipFile) {
  const responseZip = new AdmZip(zipFile)
  const docsEntries = responseZip.getEntries()
  // Validate the file paths within the "docs" folder
  const docsFilePaths = docsEntries.map((e) => e.entryName)
  const validationError = validatePluginDocsFiles(docsFilePaths)
  if (validationError) return [validationError, null]
  // If valid, filter for MDX files only, and return
  // a { filePath, fileString } object for each mdx file
  const docsMdxFiles = docsEntries
    .filter((e) => {
      return path.extname(e.entryName) === '.mdx'
    })
    .map((e) => {
      const filePath = e.entryName
      const fileString = e.getData().toString()
      return { filePath, fileString }
    })
  return [null, docsMdxFiles]
}

module.exports = fetchDevPluginDocs
