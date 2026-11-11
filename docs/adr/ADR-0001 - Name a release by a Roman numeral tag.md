---
adr-status: Accepted
superseded-by:
---

# Name a release by a Roman numeral tag

A release is named by the git tag alone, a Roman numeral without a prefix (`I`, `II`, `III`, `IV`, …), and that tag is the **Version** of both the program and the GitHub Action that installs it. The program learns its version from the tag when goreleaser builds it, so `--version` prints `IV` and no file is edited to cut a release; because `IV` is not a semantic version, the release workflow runs goreleaser with `--skip=validate`. The project has no semantic versions and nothing downstream needs them, and a numeral reads as a name, not as a promise about compatibility.

## Considered Options

- **Plain integers with a `v` prefix (`v1`, `v2`)**: rejected. goreleaser accepts them, but they look like the first part of a semantic version and invite the assumption that `v2` is a breaking change of `v1`.
- **A semantic tag under the hood (`v4.0.0`) with the numeral shown only by `--version`**: rejected. The tag, the release, and what `--version` prints would then be three different strings for one version.
- **A version constant in the source, bumped by hand**: rejected. It is the process this replaces: one more edit before every release, and a second source of truth that can disagree with the tag.

## Consequences

goreleaser refuses a tag that is not a semantic version (`failed to parse tag … as semver`) unless validation is skipped. `--skip=validate` also drops its checks of the git state, such as a dirty tree, which is harmless in CI on a clean checkout of a pushed tag but is a safety removed on purpose. Inside such a run `{{ .Version }}` is the tag itself, while `{{ .RawVersion }}` is `0.0.0` and must not be used. `go install …@latest` does not resolve these tags, since Go modules require semantic versions; the Action is the supported way to install.
