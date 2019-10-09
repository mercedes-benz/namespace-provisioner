#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
set -eo pipefail

# Get the current tag.
# If we are currently not on a tag, an empty string is returned.
currentTag() {
    local tag
    tag=$(git describe --tags --exact-match 2>/dev/null)
    echo -n "$tag"
}

# Get the last accessible tag.
lastAccessibleTag() {
    local tag
    tag=$(git describe --tags --abbrev=0 2>/dev/null)
    if [[ -z $tag ]]; then
        tag=0.0.0
    fi
    echo -n "$tag"
}

# Get the next patch version.
nextVersion() {
    local version=$1

    IFS='.' read -r -a version_array <<<"$version"
    local major=${version_array[0]}
    local minor=${version_array[1]}
    local patch=${version_array[2]}
    local next_patch=$((patch + 1))

    version="${major}.${minor}.${next_patch}"
    echo -n "$version"
}

# Echoes the current version.
#
# Versioning follows SemVer 2 (https://semver.org/), but due to
# internal limitations the build metadata separator is set to '-'.
#
# <version>:           <major>.<minor>.<patch>-<pre-release>-<build-metadata>
# <major/minor/patch>: [0-9]+
# <pre-release>:       <cleaned-branch-name>
# <build-metadata>:    <yyyymmdd>.<commit-hash>
#
# Examples:
#    release version:                               3.2.1
#    snapshot version 'develop' branch:             3.2.1-develop-20190123.58c323d
#    snapshot version 'feature/new-feature' branch: 3.2.1-new-feature-20190123.58c323d
version() {
    readonly DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    readonly SEPARATOR="-"

    # First check if we are exactly on a tag
    local version
    version=$(currentTag)
    if [[ -z $version ]]; then

        # If we are not on a tag, get version from last accessible tag.
        version=$(lastAccessibleTag)

        # Calculate next patch version.
        version=$(nextVersion "$version")

        # Get branch name
        # shellcheck source=./branch.sh
        source "$DIR/branch.sh"
        local pre_release
        pre_release=$(branch)

        if [[ "$*" == "--short" ]]; then
            # Extend version with branch name
            version="${version}-${pre_release}"
        else
            local commit_hash
            commit_hash=$(git log -n 1 --format=%h)
            # Remove leading zeros.
            commit_hash=$(echo -n "$commit_hash" | sed 's/^0*//')
            local build_metadata
            build_metadata=$(date +%Y%m%d).${commit_hash}
            # Extend version with branch name, date and commit hash
            version="${version}-${pre_release}${SEPARATOR}${build_metadata}"
        fi
    fi
    echo -n "$version"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    version "$@"
fi
