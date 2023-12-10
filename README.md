# Update Release

`gh-update-release` is a command line tool to update strings in one or more GitHub releases. This is useful if after a migration to GitHub and you want to update the download links in your release notes.

## Installation

Download the latest release from the [releases page](https://github.com/lindluni/gh-update-release/releases) page and make it executable.

## Usage

### Update all releases

```shell
gh-update-release --owner <owner> --repo <repo> --value <old_value> --replacement <new_value> --token <GitHub_PAT> --all
```

### Update a single release

```shell
gh-update-release --owner <owner> --repo <repo> --value <old_value> --replacement <new_value> --token <GitHub_PAT> --release <release_tag>
```