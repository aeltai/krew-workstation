#!/usr/bin/env bash
# Publish Krew extension to krew-workstation repo
# Run from ui-plugin-examples root: ./pkg/krew/publish-extension.sh

set -e
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BASE_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
KREW_EXT="$BASE_DIR/extensions/krew/0.1.0"

cd "$BASE_DIR"
yarn build-pkg krew

mkdir -p "$KREW_EXT/plugin"
cp dist-pkg/krew-0.1.0/krew-0.1.0.umd.min.js "$KREW_EXT/plugin/"
cp pkg/krew/extensions/krew/0.1.0/plugin/package.json "$KREW_EXT/plugin/"

echo "Extension built at: $KREW_EXT"
echo ""
echo "To push to krew-workstation:"
echo "  cd /path/to/krew-workstation"
echo "  rm -rf extensions/krew/0.1.0"
echo "  cp -r $KREW_EXT extensions/krew/"
echo "  git add extensions/"
echo "  git commit -m 'Add plugin/package.json for Rancher UIPlugin'"
echo "  git push origin main"
