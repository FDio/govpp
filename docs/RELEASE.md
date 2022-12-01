# Release

This document desribes the GoVPP releases.

## Versioning

The GoVPP uses the [Semantic Versioning](https://semver.org/) for the release versions.

The current MAJOR version of GoVPP is `0`, meaning it's still under development that might occasionally do larger breaking changes. However, we try to avoid this if possible to minimize the user impact.

## Release Cycle

The MINOR releases of GoVPP should be released **approximately 2 weeks after a VPP release** at minimum. Any additional MINOR releases for GoVPP can possibly happen between VPP releases if needed.

The PATCH releases should be released whenever it is required to publish a fix to users.

## Release Tracking

Each release has its own [milestone](https://docs.github.com/en/issues/using-labels-and-milestones-to-track-work/about-milestones) created with a due date of the expected release date. List of milestones can be found here: https://github.com/FDio/govpp/milestones

Every issue/PR that shoukd be part of a specific release should have milestone set.

## Release Process

1. Verify the state of `master` branch for the release:
  - Check if all issues/PRs that are part of the release milestone are closed
  - Check if the generated `binapi` is compatible with the lastest VPP release
2. Prepare the release in a PR with the following changes:
  - Update [CHANGELOG.md](https://github.com/FDio/govpp/blob/master/CHANGELOG.md) with the list of changes for new version
  - Update version in [version](https://github.com/FDio/govpp/blob/master/version/version.go) package to `v0.X.0` without any suffix (remove `-dev`)
3. Once PR merges to master, tag it with `v0.X.0` using annotated, signed tag & push the tag to repository
  ```sh
  git tag --sign --annotate --message "govpp v0.X.0" v0.X.0
  ```
6. After the release tag is pushed, begin development of the next release using the following steps:
  - Update version in [version](https://github.com/FDio/govpp/blob/master/version/version.go) package to `v0.X+1.0-dev`
