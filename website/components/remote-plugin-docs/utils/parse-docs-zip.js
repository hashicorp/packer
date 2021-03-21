const path = require('path')
const AdmZip = require('adm-zip')
const validatePluginDocsFiles = require('./validate-plugin-docs-files')

/*

NOTE: used for default `docs.zip` release assets

*/

// Given a response from fetching a docs.zip file,
// which is a compressed "docs" folder,
//
// return [null, docsMdxFiles] if docs files
// are successfully fetched and valid,
// where docsMdxFiles is an array of { filePath, fileString } items.
//
// otherwise, return [err, null]
// where err is an error message describing whether the
// docs files were missing or invalid, with a path to resolution
async function parseDocsZip(response) {
  // the file path from the repo root is the same as the zip entryName,
  // which includes the docs directory as the first part of the path
  const responseBuffer = Buffer.from(await response.arrayBuffer())
  const responseZip = new AdmZip(responseBuffer)
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

/*
   const dirs = path.dirname(e.entryName).split('/')
      const pathFromDocsDir = dirs.slice(1).join('/')
      */

module.exports = parseDocsZip
