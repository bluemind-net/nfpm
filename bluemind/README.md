# nfpm for BlueMind

BlueMind packages are built with this fork of [nfpm](https://github.com/goreleaser/nfpm): the
bm-packaging-maven-plugin resolves `net.bluemind:nfpm:<version>:tar.gz:<os>-<arch>` from the BlueMind Nexus and
uses it for the deb and the rpm packages.

The stock releases cannot build RPMs that behave like the ones rpmbuild built from the former `.spec` files:
scriptlet interpreters are hard-coded to `/bin/sh` and RPM triggers are not supported.

## Content of the `bluemind` branch

Upstream `main`, plus a merge of each change BlueMind needs. Each change is also proposed upstream, from its own
branch of this fork:

| Change | Upstream PR | Branch |
|---|---|---|
| `rpm.triggers` | [#1118](https://github.com/goreleaser/nfpm/pull/1118), with the review fixes | `rpm-triggers-fixes` |
| `rpm.interpreters`, scriptlets require their interpreter | [#1149](https://github.com/goreleaser/nfpm/pull/1149) | `rpm-scriptlet-interpreters` |
| `rpm.doc_dirs` | [#1150](https://github.com/goreleaser/nfpm/pull/1150) | `rpm-doc-dirs` |
| `rpm.script_requires` | [#1151](https://github.com/goreleaser/nfpm/pull/1151) | `rpm-script-requires` |
| ELF file colors | [#1152](https://github.com/goreleaser/nfpm/pull/1152) | `rpm-elf-colors` |
| `go.digitalxero.dev/rpm` v0.3.0: `name(arch)` provide | [#1141](https://github.com/goreleaser/nfpm/pull/1141) | upstream PR |

The last commit adds this directory, `.gitlab-ci.yml` and the `.gitignore` entry.

When all these PRs are merged and released, drop this fork and use the upstream release.

## Release

1. Bump `<version>` in `pom.xml`: `<nfpm release line>.<upstream main commit>-bm<n>`.
2. Push the branch: the CI tests, runs `build.sh` (Linux and macOS, amd64 and arm64) and deploys to the Nexus.
3. Bump the version in `Nfpm.java` of bm-packaging-maven-plugin.

Update to a new upstream: recreate the branch from upstream `main`, merge the branches of the table again (drop the
ones merged upstream), add the packaging commit, run `go test ./...`, then release.
