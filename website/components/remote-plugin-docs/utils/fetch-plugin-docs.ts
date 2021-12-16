import fetch from 'isomorphic-unfetch'
import parseSourceZip from './parse-source-zip'
import parseDocsZip from './parse-docs-zip'

// Given a repo and tag,
//
// return [null, docsMdxFiles] if docs files
// are successfully fetched and valid,
// where docsMdxFiles is an array of { filePath, fileString } items.
//
// otherwise, return [err, null]
// where err is an error message describing whether the
// docs files were missing or invalid, with a path to resolution
async function fetchDocsFiles({ repo, tag }: { repo: string; tag: string }) {
  // If there's a docs.zip asset, we'll prefer that
  const docsZipUrl =
    tag === 'latest'
      ? `https://github.com/${repo}/releases/latest/download/docs.zip`
      : `https://github.com/${repo}/releases/download/${tag}/docs.zip`
  const docsZipResponse = await fetch(docsZipUrl, { method: 'GET' })
  const hasDocsZip = docsZipResponse.status === 200
  // Note: early return!
  if (hasDocsZip) return await parseDocsZip(docsZipResponse)
  // Else if docs.zip is not present, and we only have the "latest" tag,
  // then throw an error - we can't resolve the fallback source ZIP
  // unless we resort to calling the GitHub API, which we do not want to do
  if (tag === 'latest') {
    const err = `Failed to fetch. Could not find "docs.zip" at ${docsZipUrl}. To fall back to parsing docs from "source", please provide a specific version tag instead of "${tag}".`
    return [err, null] as const
  }
  // Else if docs.zip is not present, and we have a specific tag, then
  // fall back to parsing docs files from the source zip
  const sourceZipUrl = `https://github.com/${repo}/archive/${tag}.zip`
  const sourceZipResponse = await fetch(sourceZipUrl, { method: 'GET' })
  const missingSourceZip = sourceZipResponse.status !== 200
  if (missingSourceZip) {
    const err = `Failed to fetch. Could not find "docs.zip" at ${docsZipUrl}, and could not find fallback source ZIP at ${sourceZipUrl}. Please ensure one of these assets is available.`
    return [err, null] as const
  }
  // Handle parsing from plugin source zip
  return await parseSourceZip(sourceZipResponse)
}

async function fetchPluginDocs({ repo, tag }) {
  const [err, docsMdxFiles] = await fetchDocsFiles({ repo, tag })
  if (err) {
    const errMsg = `Invalid plugin docs ${repo}, on release ${tag}. ${err}`
    throw new Error(errMsg)
  }
  return docsMdxFiles
}

function memoize(method) {
  let cache = {}
  return async function () {
    let args = JSON.stringify(arguments)
    if (!cache[args]) {
      cache[args] = method.apply(this, arguments)
    }
    return cache[args]
  }
}

export default memoize(fetchPluginDocs)
