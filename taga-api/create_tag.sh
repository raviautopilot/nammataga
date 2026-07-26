#!/usr/bin/env bash
set -euo pipefail

# create_tag.sh
# 1) Ensure we're on main and up-to-date
# 2) Run go tests
# 3) Create next patch tag (example: v0.1.1 -> v0.1.2) and push it

die() { echo "ERROR: $*" >&2; exit 1; }

# Ensure git repo
git rev-parse --is-inside-work-tree >/dev/null 2>&1 || die "Not inside a git repository."

# Ensure clean working tree (recommended)
if ! git diff --quiet || ! git diff --cached --quiet; then
  die "Working tree is not clean. Commit/stash your changes before tagging."
fi

# Ensure branch is main
branch="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$branch" != "main" ]]; then
  die "You must be on branch 'main' (current: $branch)"
fi

# Pull latest
echo "Pulling latest from main..."
git pull --ff-only

# Run tests
echo "Running tests..."
go test ./...

# Find latest semver tag like vX.Y.Z (ignore non-matching tags)
latest_tag="$(git tag --list 'v*' --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -n 1 || true)"
if [[ -z "${latest_tag}" ]]; then
  die "No existing semver tag found (expected something like v0.1.1). Create the first tag manually (e.g., git tag v0.1.0)."
fi

# Compute next patch (minor stays the same, patch +1)
ver="${latest_tag#v}"
IFS='.' read -r major minor patch <<<"$ver"

[[ "${major:-}" =~ ^[0-9]+$ ]] || die "Invalid major in tag: $latest_tag"
[[ "${minor:-}" =~ ^[0-9]+$ ]] || die "Invalid minor in tag: $latest_tag"
[[ "${patch:-}" =~ ^[0-9]+$ ]] || die "Invalid patch in tag: $latest_tag"

new_patch=$((patch + 1))
new_tag="v${major}.${minor}.${new_patch}"

# Ensure tag doesn't already exist
if git rev-parse "$new_tag" >/dev/null 2>&1; then
  die "Tag already exists: $new_tag"
fi

# Create annotated tag
echo "Creating tag: $new_tag"
git tag -a "$new_tag" -m "Release $new_tag"

# Push the tag to remote
echo "Pushing tag to origin..."
git push origin "$new_tag"

echo "✓ Successfully created and pushed tag: $new_tag (from $latest_tag)"