#!/usr/bin/env bash
# SPDX-License-Identifier: MIT
set -eo pipefail

# Echoes the cleaned name of the current branch
# e.g.
# master => master
# develop => develop
# feature/new-feature => new-feature
# poc/other-feature => poc-other-feature
branch() {
    # Try to get branch name from Git
    local branch_name
    branch_name=$(git rev-parse --abbrev-ref HEAD)
    if [[ $branch_name == "HEAD" ]]; then
        # Try to get branch name from GitHub Action environment variable
        # shellcheck disable=SC2153
        if [[ -n $GITHUB_REF ]]; then
            branch_name=$(echo -n "$GITHUB_REF" | sed 's#refs/heads/##')
        # Try to get branch name from Jenkins environment variable
        elif [[ -n $BRANCH_NAME ]]; then
            branch_name=$BRANCH_NAME
        else
            echo "error: detached HEAD" >&2
            exit 1
        fi
    fi

    # Clean branch name
    if [[ $branch_name =~ ^feature ]]; then
        branch_name=${branch_name//feature\//}
    fi
    branch_name=${branch_name//[^-a-zA-Z0-9]/-}
    branch_name=$(echo -n "$branch_name" | tr '[:upper:]' '[:lower:]')

    echo -n "${branch_name}"
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    branch "$@"
fi
