# nfpm for BlueMind

BlueMind packages are built with this fork of [nfpm](https://github.com/goreleaser/nfpm). The releases of this fork
carry a Linux amd64 binary, which the bm-packaging-maven-plugin uses for the deb and the rpm packages.

The stock releases cannot build RPMs that behave like the ones rpmbuild built from the former `.spec` files:
scriptlet interpreters are hard-coded to `/bin/sh` and RPM triggers are not supported.

## Content of the `bluemind` branch

Upstream `main`, plus a merge of each change BlueMind needs. Each change is also proposed upstream, from its own
branch of this fork:

| Change | Upstream PR | Branch |
|---|---|---|
| `rpm.triggers` | [#1118](https://github.com/goreleaser/nfpm/pull/1118), with the review fixes of [#1153](https://github.com/goreleaser/nfpm/pull/1153) | `rpm-triggers-fixes` |
| `rpm.interpreters`, scriptlets require their interpreter | [#1149](https://github.com/goreleaser/nfpm/pull/1149) | `rpm-scriptlet-interpreters` |
| `rpm.doc_dirs` | [#1150](https://github.com/goreleaser/nfpm/pull/1150) | `rpm-doc-dirs` |
| `rpm.script_requires` | [#1151](https://github.com/goreleaser/nfpm/pull/1151) | `rpm-script-requires` |

The last commit adds this directory, `.github/workflows/bluemind-release.yml` and the `.gitignore` entry.

When all these PRs are merged and released, drop this fork and use the upstream release.

## Release

1. Run `go test ./...`: the release workflow does not run the tests.
2. Tag the branch with `<nfpm release line>.<upstream main commit>-bm<n>`, for example `2.48.0.37cd27d-bm1`, and
   push the tag. The `bluemind-release` workflow builds `nfpm-<tag>-linux-amd64.tar.gz` and attaches it to the
   GitHub release of the tag. The tag must not start with `v`: the upstream release workflow runs on `v*` tags.
3. Bump the version in `Nfpm.java` of bm-packaging-maven-plugin.

To build the binary locally: `bluemind/build.sh <version>`.

Update to a new upstream: recreate the branch from upstream `main`, merge the branches of the table again (drop the
ones merged upstream), add the packaging commit, then release.
