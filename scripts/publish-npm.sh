#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT_DIR"

if ! npm whoami >/dev/null 2>&1; then
  echo "Not logged in to npm. Run: npm adduser"
  exit 1
fi

npm run release:prepare
npm publish --access public
