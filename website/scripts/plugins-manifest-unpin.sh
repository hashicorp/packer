# Get the count of entries in the plugin manifest
ENTRY_COUNT=$(jq -r ".|length" ./data/plugins-manifest.json)
# For each plugin manifest entry, change `version` to `"latest"`.
# Uses `jq` - https://formulae.brew.sh/formula/jq
for ((i = 0; i < ENTRY_COUNT; i++)); do
  PLUGIN_REPO=$(jq -r ".[$i].repo" ./data/plugins-manifest.json)
  echo "Setting \"$PLUGIN_REPO\" to version \"latest\"..."
  cat <<<$(jq ".[$i].version = \"latest\"" ./data/plugins-manifest.json) >./data/plugins-manifest.json
done
