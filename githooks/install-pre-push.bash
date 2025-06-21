#!/bin/bash

# Print a message and kill the script.
die() {
    echo "$@" 1>&2
    exit 1
}

# Finds the top of the repo.
find_git_repo_top() {
    local current_dir=$(pwd)

    # Loop until reaching the root directory "/"
    while [ "$current_dir" != "/" ]; do
        # Check if ".git" directory exists
        if [ -d "$current_dir/.git" ]; then
            echo "$current_dir"
            return
        fi

        # Move up one directory
        current_dir=$(dirname "$current_dir")
    done

    # If ".git" directory is not found
    echo "Git repository not found."
    exit 1
}

# Ask the user a yes/no question and await their response. Return 0 if
# they say yes (in some format).
await_yes_no() {
    read -r answer
    case "$answer" in
        [yY]|[yY][eE][sS])
            echo 0
            ;;
        *)
            echo 1
            ;;
    esac
}

# Delete whatever hooks may be active in the .git/hooks directory. This
# may include things like the old pre-commit hook we had been using
# prior for April, 2024.
delete_existing_hooks_with_confirmation() {
    project_root="$1"
    hooks=$(find "${project_root}/.git/hooks/" -mindepth 1 ! -name "*.sample")
    echo "Found hook files: $hooks"
    echo "OK to delete? [Y/n]"
    if [ "$(await_yes_no)" -ne 0 ]; then
        die "OK; aborting."
    fi
    rm -f -r $hooks
} 

# Install the pre-push script and any hooks found under the ./scripts
# directory.
install_pre_push_hooks() {
    project_root="$1"
    echo "Installing scripts into .git/hooks ..."
    mkdir -p "${project_root}/.git/hooks"
    ln -s "${project_root}/scripts/pre-push" "${project_root}/.git/hooks/pre-push"
}

# Installs pre-push scripts after ensuring we're running from the
# directory root, and after cleaning up any old git hook scripts.
main() {
    project_root="$(find_git_repo_top)"
    delete_existing_hooks_with_confirmation "$project_root"
    install_pre_push_hooks "$project_root"
}

main