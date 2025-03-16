# Release Process

This document describes the steps to release a new version of otel-tui.

## Background

This project uses Go workspaces and includes a submodule `tuiexporter`. To make the project installable via `go install`, the `tuiexporter` module is referenced by commit hash in the main module's `go.mod`. Therefore, the release process requires specific steps to ensure proper versioning.

## Steps

1. Run the "prepare-release" workflow on GitHub Actions

   - This workflow updates the `tuiexporter` version in `go.mod` to reference the latest commit
   - A Pull Request will be automatically created with the necessary changes

2. Review and merge the created Pull Request

   - The PR will update `go.mod` and `go.sum`
   - Ensure the changes are correct before merging

3. Create a new release on GitHub

   - Create a new tag (e.g., `v1.2.3`)
   - Write release notes documenting the changes
   - Publish the release

4. Merge the Nix update PR
   - After the release is published, a PR to update the Nix package will be created
   - Review and merge this PR to update the package in the Nix ecosystem
