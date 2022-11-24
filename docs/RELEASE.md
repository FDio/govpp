# Release

This document desribes the release process for GoVPP.

## Release Cycle

The GoVPP releases should happen approximately 2 weeks after VPP release at minimum. Additional releases can happen between VPP releases if needed.

Patch version releases can happen whenenever required.

## Milestones

Each release has it's own milestone created with approximate release date. List of milestones can be found here: https://github.com/FDio/govpp/milestones

Every issue/PR that shoukd be part of a specific release should have milestone set.

## Release Process

1. Check if all issues/PRs that are part of the release are closed
2. Update [CHANGELOG.md](https://github.com/FDio/govpp/blob/master/CHANGELOG.md)
3. Update version in [version](https://github.com/FDio/govpp/blob/master/version/version.go) package
4. Create new version tag & push it
5. Update version to next version in development
