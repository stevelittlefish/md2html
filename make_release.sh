#!/usr/bin/env bash
#
# Cut a release by pushing a version tag.
#
# The Build workflow triggers on pushed v* tags: it tests, builds the Linux and
# Windows binaries, packages the macOS source archive with install.sh, creates a
# GitHub Release named after the tag, and attaches the downloads. This script
# just creates and pushes the tag.
#
# Usage:
#   ./make_release.sh            Show the latest release tag and the suggested next one.
#   ./make_release.sh vX.Y.Z     Create and push that tag, cutting the release.
#   ./make_release.sh X.Y.Z      Same; a leading "v" is added for you.

set -euo pipefail

cd "$(git rev-parse --show-toplevel)"

latest_tag() {
	git tag --list 'v*' --sort=-version:refname | head -n1
}

# Given vX.Y.Z, print the next patch version.
suggest_next() {
	local major minor patch
	IFS=. read -r major minor patch <<<"${1#v}"
	if [[ ! $major =~ ^[0-9]+$ || ! $minor =~ ^[0-9]+$ || ! $patch =~ ^[0-9]+$ ]]; then
		echo "v0.1.0"
		return
	fi
	echo "v${major}.${minor}.$((patch + 1))"
}

if [[ $# -eq 0 ]]; then
	current="$(latest_tag)"
	if [[ -z $current ]]; then
		echo "No release tags yet."
		echo "Suggested first release: v0.1.0"
		echo
		echo "Cut it with: ./make_release.sh v0.1.0"
	else
		next="$(suggest_next "$current")"
		echo "Latest release tag: $current"
		echo "Suggested next tag: $next"
		echo
		echo "Cut it with: ./make_release.sh $next"
	fi
	exit 0
fi

if [[ $# -gt 1 ]]; then
	echo "error: expected a single version argument (e.g. v1.2.3)" >&2
	exit 2
fi

version="$1"
[[ $version == v* ]] || version="v$version"

if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
	echo "error: '$version' is not a valid version tag (expected vX.Y.Z)" >&2
	exit 2
fi

if git rev-parse -q --verify "refs/tags/$version" >/dev/null; then
	echo "error: tag $version already exists" >&2
	exit 1
fi

branch="$(git rev-parse --abbrev-ref HEAD)"
if [[ $branch != main ]]; then
	echo "warning: you are on '$branch', not 'main'." >&2
fi

if [[ -n "$(git status --porcelain)" ]]; then
	echo "warning: working tree has uncommitted changes; the tag will point at the last commit." >&2
fi

if [[ -n "$(git log --oneline '@{u}..' 2>/dev/null)" ]]; then
	echo "warning: main has commits that are not pushed; pushing them first." >&2
	git push origin HEAD
fi

echo "Tagging $(git rev-parse --short HEAD) as $version and pushing to origin..."
git tag -a "$version" -m "Release $version"
git push origin "$version"

echo
echo "Pushed $version. Watch the release build here:"
echo "  https://github.com/stevelittlefish/md2html/actions/workflows/build.yml"
echo "The release will appear at:"
echo "  https://github.com/stevelittlefish/md2html/releases/tag/$version"
