#!/usr/bin/env bash
set -euo pipefail

# Compatibility wrapper. New documentation and automation should use batch.sh.
script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
exec "$script_dir/batch.sh" "$@"
