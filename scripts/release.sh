#!/usr/bin/env bash
# Creates a signed vX.Y.Z tag for the next semver release and optionally pushes it.
set -euo pipefail

branch="${RELEASE_BRANCH:-main}"
remote="${RELEASE_REMOTE:-origin}"

bold=$'\e[1m' dim=$'\e[2m' red=$'\e[31m' green=$'\e[32m' yellow=$'\e[33m' blue=$'\e[34m' magenta=$'\e[35m' reset=$'\e[0m'

die() { echo "${red}✗${reset} $*" >&2; exit 1; }
ok() { echo "${green}✓${reset} $*"; }
step() { echo "${dim}…${reset} $*"; }

command -v fzf >/dev/null || die "fzf is required: brew install fzf"

pick() {
  local label=$1 header=$2
  shift 2
  printf '%s\n' "$@" | fzf --ansi --no-input --no-sort --layout=reverse --height=~20 \
    --border=rounded --border-label=" $label " --header="$header" --header-first \
    --pointer='▶' --color='label:bold,pointer:yellow,header:dim' --delimiter=$'\t' --with-nth=2 |
    cut -f1
}

cd "$(git rev-parse --show-toplevel)"

[ "$(git branch --show-current)" = "$branch" ] || die "not on $branch"
[ -z "$(git status --porcelain)" ] || die "working tree is not clean"

step "fetching $remote/$branch"
git fetch --quiet --tags "$remote" "$branch" || die "fetch from $remote failed"
[ "$(git rev-parse HEAD)" = "$(git rev-parse "$remote/$branch")" ] || die "$branch is not in sync with $remote/$branch"
ok "$branch is clean and in sync with $remote"

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

commits=$(git log --format="%C(yellow)%h%C(reset) %s" --color=always "$range")
count=$(git rev-list --count "$range")
current="${latest:-v0.0.0}"

echo
echo "${bold}Yubi Ki-Yay release${reset}  ${dim}current${reset} ${magenta}${latest:-none}${reset}  ${dim}·${reset} $count commit(s) since"
echo "$commits" | sed 's/^/  /'
echo

next_for() {
  case $1 in
    major) echo "v$((major + 1)).0.0" ;;
    minor) echo "v$major.$((minor + 1)).0" ;;
    patch) echo "v$major.$minor.$((patch + 1))" ;;
  esac
}

options=()
for kind in "$suggest" major minor patch; do
  [ "$kind" = "$suggest" ] && [ ${#options[@]} -gt 0 ] && continue
  label=$(printf '%-6s %s → %s%s' "$kind" "$current" "${bold}$(next_for "$kind")${reset}" \
    "$([ "$kind" = "$suggest" ] && echo "  ${green}suggested${reset}")")
  options+=("$kind"$'\t'"$label")
done

bump=$(pick "Bump" "↑↓ choose · enter select · esc cancel" "${options[@]}") || die "cancelled"
tag=$(next_for "$bump")

git rev-parse -q --verify "refs/tags/$tag" >/dev/null && die "$tag already exists"

action=$(pick "$tag" "$current → $tag" \
  "push"$'\t'"Create signed tag and push ${dim}(starts the release)${reset}" \
  "local"$'\t'"Create signed tag only" \
  "cancel"$'\t'"${red}Cancel${reset}") || die "cancelled"
[ "$action" != cancel ] || die "cancelled"

git tag -s "$tag" -m "Yubi Ki-Yay ${tag#v}"
ok "created signed tag ${bold}$tag${reset}"

if [ "$action" = push ]; then
  step "pushing $tag to $remote"
  git push --quiet "$remote" "refs/tags/$tag"
  ok "pushed ${bold}$tag${reset} ${blue}→ release workflow started${reset}"
else
  echo "${dim}push later with:${reset} git push $remote $tag"
fi
