# Release

This document desribes the release process for GoVPP.

## Release Cycle

The GoVPP releases should happen approximately 2 weeks after VPP release at minimum. Additional releases can happen between VPP releases if needed.

Patch version releases can happen whenenever required.

## Milestones

Each release has it's own milestone created with approximate release date. List of milestones can be found here: https://github.com/FDio/govpp/milestones

Every issue/PR that shoukd be part of a specific release should have milestone set.

## Release Process

1. Check if all issues/PRs that are part of the release milestone are closed
2. Open new PR with the following changes:
  - Update [CHANGELOG.md](https://github.com/FDio/govpp/blob/master/CHANGELOG.md) with the list of changes for new version
  - Update version in [version](https://github.com/FDio/govpp/blob/master/version/version.go) package to `v0.X.0` without any suffix (remove `-dev`)
3. Once PR merges to master, tag it with `v0.X.0` using annotted signed tag & push the tag to repo
  ```sh
  git tag --sign --annotate --message "govpp v0.X.0" v0.X.0
  ```
6. Once the tag is pushed, do another commit to master to start development of next version
  - Update version in [version](https://github.com/FDio/govpp/blob/master/version/version.go) package to `v0.X+1.0-dev`
