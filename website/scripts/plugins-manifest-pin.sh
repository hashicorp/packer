# Get the count of entries in the plugin manifest
ENTRY_COUNT=$(jq -r ".|length" ./data/plugins-manifest.json)
# For each plugin manifest entry, pin to the latest version.
# Uses two utilities:
# - GitHub CLI - https://formulae.brew.sh/formula/gh
# - `jq` - https://formulae.brew.sh/formula/jq
for ((i = 0; i < ENTRY_COUNT; i++)); do
  PLUGIN_REPO=$(jq -r ".[$i].repo" ./data/plugins-manifest.json)
  API_URL="/repos/$PLUGIN_REPO/releases/latest"
  PINNED_VERSION=$(gh api -H "Accept: application/vnd.github+json" $API_URL | jq -r '.tag_name')
  echo "Pinning \"$PLUGIN_REPO\" to version \"$PINNED_VERSION\"..."
  cat <<<$(jq ".[$i].version = \"$PINNED_VERSION\"" ./data/plugins-manifest.json) >./data/plugins-manifest.json
done
