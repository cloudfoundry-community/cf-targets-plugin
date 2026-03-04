#!/usr/bin/env bash
#
# Check whether the vendored internal/diff package has upstream changes
# since the last sync point recorded in internal/diff/SYNC_VERSION.
#
# Exit codes:
#   0 — no upstream changes detected (in sync)
#   1 — upstream changes detected
#   2 — script error (missing tools, network failure, etc.)
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
SYNC_VERSION_FILE="${PROJECT_ROOT}/internal/diff/SYNC_VERSION"
UPSTREAM_REPO="https://github.com/golang/tools.git"
UPSTREAM_PATH="internal/diff"

# Files we track from upstream (relative to internal/diff/ in the upstream repo).
# Excludes our local additions: LICENSE, SYNC_VERSION, git.sh, and test files.
TRACKED_FILES=(
    "diff.go"
    "merge.go"
    "ndiff.go"
    "unified.go"
    "lcs/common.go"
    "lcs/doc.go"
    "lcs/labels.go"
    "lcs/old.go"
    "lcs/sequence.go"
)

# --- helpers ----------------------------------------------------------------

die() {
    echo "ERROR: $*" >&2
    exit 2
}

# --- preflight --------------------------------------------------------------

command -v git >/dev/null 2>&1 || die "git is not installed"

[[ -f "${SYNC_VERSION_FILE}" ]] || die "SYNC_VERSION file not found at ${SYNC_VERSION_FILE}"

SYNC_TAG="$(tr -d '[:space:]' < "${SYNC_VERSION_FILE}")"
[[ -n "${SYNC_TAG}" ]] || die "SYNC_VERSION file is empty"

echo "=== Upstream Diff Sync Check ==="
echo ""
echo "  Sync version : ${SYNC_TAG}"
echo "  Upstream repo : ${UPSTREAM_REPO}"
echo "  Upstream path : ${UPSTREAM_PATH}/"
echo ""

# --- clone upstream (treeless — full history, no blobs) ---------------------

TMPDIR_ROOT="$(mktemp -d)"
trap 'rm -rf "${TMPDIR_ROOT}"' EXIT

echo "Cloning upstream (treeless)..."
if ! git clone --filter=blob:none --quiet "${UPSTREAM_REPO}" "${TMPDIR_ROOT}/tools" 2>/dev/null; then
    die "Failed to clone upstream repository (network issue?)"
fi
echo ""

TOOLS_DIR="${TMPDIR_ROOT}/tools"

# Verify the sync tag exists upstream
if ! git -C "${TOOLS_DIR}" rev-parse --verify --quiet "${SYNC_TAG}" >/dev/null 2>&1; then
    die "Sync tag '${SYNC_TAG}' not found in upstream repository"
fi

# --- find latest upstream tag -----------------------------------------------

LATEST_TAG="$(git -C "${TOOLS_DIR}" describe --tags --abbrev=0 HEAD 2>/dev/null || echo "unknown")"
echo "  Latest upstream tag : ${LATEST_TAG}"

if [[ "${LATEST_TAG}" == "${SYNC_TAG}" ]]; then
    echo "  Status              : UP TO DATE"
    echo ""
    echo "No upstream changes since ${SYNC_TAG}."
    exit 0
fi

# --- count commits touching internal/diff/ ---------------------------------

COMMIT_LOG="$(git -C "${TOOLS_DIR}" log --oneline "${SYNC_TAG}..HEAD" -- "${UPSTREAM_PATH}/" 2>/dev/null || true)"
COMMIT_COUNT=0
if [[ -n "${COMMIT_LOG}" ]]; then
    COMMIT_COUNT="$(echo "${COMMIT_LOG}" | wc -l | tr -d ' ')"
fi

echo "  Commits since sync  : ${COMMIT_COUNT}"
echo ""

if [[ "${COMMIT_COUNT}" -eq 0 ]]; then
    echo "No changes to ${UPSTREAM_PATH}/ since ${SYNC_TAG} (upstream is at ${LATEST_TAG})."
    exit 0
fi

# --- list changed files -----------------------------------------------------

CHANGED_FILES="$(git -C "${TOOLS_DIR}" diff --name-only "${SYNC_TAG}..HEAD" -- "${UPSTREAM_PATH}/" 2>/dev/null || true)"

echo "--- Commits ---"
echo "${COMMIT_LOG}"
echo ""

# --- cross-reference against tracked files ----------------------------------

echo "--- Changed Files (tracked by this project) ---"
HAS_TRACKED_CHANGES=false

for changed in ${CHANGED_FILES}; do
    # Strip the upstream path prefix to get the relative filename
    relative="${changed#${UPSTREAM_PATH}/}"
    for tracked in "${TRACKED_FILES[@]}"; do
        if [[ "${relative}" == "${tracked}" ]]; then
            echo "  * ${relative}"
            HAS_TRACKED_CHANGES=true
            break
        fi
    done
done

echo ""
echo "--- Changed Files (not tracked / upstream-only) ---"
for changed in ${CHANGED_FILES}; do
    relative="${changed#${UPSTREAM_PATH}/}"
    is_tracked=false
    for tracked in "${TRACKED_FILES[@]}"; do
        if [[ "${relative}" == "${tracked}" ]]; then
            is_tracked=true
            break
        fi
    done
    if [[ "${is_tracked}" == "false" ]]; then
        echo "  - ${relative}"
    fi
done

echo ""
echo "=== Verdict ==="
if [[ "${HAS_TRACKED_CHANGES}" == "true" ]]; then
    echo "UPSTREAM CHANGES DETECTED in tracked files."
    echo "Review the commits above and consider syncing from ${SYNC_TAG} to ${LATEST_TAG}."
    exit 1
else
    echo "Upstream has new commits but NONE affect tracked files."
    echo "No sync action required (upstream: ${LATEST_TAG}, synced: ${SYNC_TAG})."
    exit 0
fi
