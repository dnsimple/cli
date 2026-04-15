# Releasing

This document describes the steps to release a new version of DNSimple/CLI.

## Prerequisites

- You have commit access to the repository
- You have push access to the repository
- You have a GPG key configured for signing tags
- The repository GitHub Actions secrets required for release publishing are configured

## Release process

1. **Determine the new version** using [Semantic Versioning](https://semver.org/)

   ```shell
   VERSION=X.Y.Z
   ```

   - **MAJOR** version for incompatible API changes
   - **MINOR** version for backwards-compatible functionality additions
   - **PATCH** version for backwards-compatible bug fixes

2. **Run tests** and confirm they pass

   ```shell
   go test -v ./...
   ```

3. **Update the changelog** with the new version

   Finalize the `## Unreleased` section in `CHANGELOG.md` assigning the version.

4. **Commit the new version**

   ```shell
   git commit -a -m "Release $VERSION"
   ```

5. **Push the changes**

   ```shell
   git push origin main
   ```

6. **Wait for CI to complete**

7. **Create a signed tag**

   ```shell
   git tag -a v$VERSION -s -m "Release $VERSION"
   git push origin --tags
   ```

## Post-release

- Verify the GitHub release was created in `dnsimple/dnsimple-cli`
- Verify the public release mirror was updated in `dnsimple/homebrew-tap`
- Verify the Homebrew formula was updated in `dnsimple/homebrew-tap`
- Announce the release if necessary
