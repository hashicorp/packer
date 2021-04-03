const path = require('path')

const COMPONENT_TYPES = [
  'builders',
  'datasources',
  'post-processors',
  'provisioners',
]

// Given an array of file paths within the "docs" folder,
// validate that no unexpected files are being included,
// and that there is at least one component subfolder
// with at least one .mdx file within it.
function validatePluginDocsFiles(filePaths) {
  function isValidPath(filePath) {
    // Allow the docs root folder
    const isDocsRoot = filePath === 'docs/'
    // Allow component folders
    const isComponentRoot = COMPONENT_TYPES.reduce((acc, type) => {
      return acc || filePath === `docs/${type}/`
    }, false)
    // Allow .mdx files in component folders
    const isComponentMdx = COMPONENT_TYPES.reduce((acc, type) => {
      const mdxPathRegex = new RegExp(`^docs/${type}/(.*).mdx$`)
      return acc || mdxPathRegex.test(filePath)
    }, false)
    // Allow docs/README.md files
    const isDocsReadme = filePath == 'docs/README.md'
    // Combine all allowed types
    const isValidPath =
      isDocsRoot || isComponentRoot || isComponentMdx || isDocsReadme
    return isValidPath
  }
  const invalidPaths = filePaths.filter((f) => !isValidPath(f))
  if (invalidPaths.length > 0) {
    return `Found invalid files or folders in the docs directory: ${JSON.stringify(
      invalidPaths
    )}. Please ensure the docs folder contains only component subfolders and .mdx files within those subfolders. Valid component types are: ${JSON.stringify(
      COMPONENT_TYPES
    )}.`
  }
  const validPaths = filePaths.filter(isValidPath)
  const mdxFiles = validPaths.filter((fp) => path.extname(fp) === '.mdx')
  const isMissingDocs = mdxFiles.length == 0
  if (isMissingDocs) {
    return `Could not find valid .mdx files. Please ensure there is at least one component subfolder in the docs directory, which contains at least one .mdx file. Valid component types are: ${JSON.stringify(
      COMPONENT_TYPES
    )}.`
  }
  return null
}

module.exports = validatePluginDocsFiles
