#!/usr/bin/env bash
# Creates a signed vX.Y.Z tag for the next semver release and optionally pushes it.
set -euo pipefail

branch="${RELEASE_BRANCH:-main}"
remote="${RELEASE_REMOTE:-origin}"

die() { echo "release: $*" >&2; exit 1; }

cd "$(git rev-parse --show-toplevel)"

[ "$(git branch --show-current)" = "$branch" ] || die "not on $branch"
[ -z "$(git status --porcelain)" ] || die "working tree is not clean"

git fetch --quiet --tags "$remote" "$branch" || die "fetch from $remote failed"
[ "$(git rev-parse HEAD)" = "$(git rev-parse "$remote/$branch")" ] || die "$branch is not in sync with $remote/$branch"

latest=$(git tag --list 'v*' --sort=-v:refname | grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | head -n1 || true)

if [ -n "$latest" ]; then
  range="$latest..HEAD"
  [ -n "$(git rev-list "$range")" ] || die "no commits since $latest"
  IFS=. read -r major minor patch <<<"${latest#v}"
else
  range="HEAD"
  major=0 minor=0 patch=0
fi

# Conventional Commits drive the suggestion: breaking -> major, feat -> minor, else patch.
log=$(git log --format='%s%n%b' "$range")
if [ -z "$latest" ]; then
  suggest=minor
elif grep -qE '^[a-z]+(\([^)]*\))?!:|^BREAKING[ -]CHANGE:' <<<"$log"; then
  suggest=major
elif grep -qE '^feat(\([^)]*\))?:' <<<"$log"; then
  suggest=minor
else
  suggest=patch
fi

echo "Commits since ${latest:-the start}:"
git log --oneline "$range" | sed 's/^/  /'
echo

read -r -p "Bump major (M), minor (m) or patch (p)? [default: $suggest] " answer
case "${answer:-$suggest}" in
  M|major) bump=major ;;
  m|minor) bump=minor ;;
  p|patch) bump=patch ;;
  *) die "unknown choice: $answer" ;;
esac

case "$bump" in
  major) major=$((major + 1)) minor=0 patch=0 ;;
  minor) minor=$((minor + 1)) patch=0 ;;
  patch) patch=$((patch + 1)) ;;
esac
tag="v$major.$minor.$patch"

git rev-parse -q --verify "refs/tags/$tag" >/dev/null && die "$tag already exists"

read -r -p "Create signed tag $tag (${latest:-none} -> $tag)? [y/N] " ok
[[ "$ok" =~ ^[yY]$ ]] || die "aborted"
git tag -s "$tag" -m "Yubi Ki-Yay ${tag#v}"
echo "Created $tag"

read -r -p "Push $tag to $remote and start the release? [y/N] " ok
if [[ "$ok" =~ ^[yY]$ ]]; then
  git push "$remote" "refs/tags/$tag"
else
  echo "Not pushed. Later: git push $remote $tag"
fi
