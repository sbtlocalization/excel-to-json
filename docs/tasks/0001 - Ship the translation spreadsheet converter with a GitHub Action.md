---
epic:
parent:
status: In progress
priority: 5
blocked by:
kind: Task
mode: AFK
tier: mid
---

# Sub-tasks

<!-- live the following block unchanged -->
```base
summaries: {}
filters:
  and:
    - parent == [this.file]
formulas:
  is_blocked: |-
    note["blocked by"].reduce(
      value.asFile().properties.status != 'Done' || acc,
      false
    )
  progress_icon: |-
    if(status == "Done",
      icon('square-check-big'),
      if(formula.is_blocked,
        icon('construction'), 
        icon('square')
      )
    )
  progress_sort: |-
    if(status == "Done", 
      20, 
      if(formula.is_blocked, 
        10, 
        if(status == 'In progress', 
          5, 
          0)))
  progress_string: |-
    if(status == "Done", 
      "Done", 
      if(formula.is_blocked, 
        "On hold",
        "Ready"))
  priority: if(priority, icon('tally-' + priority), null)
properties:
  note.status:
    displayName: status
  file.name:
    displayName: task
  formula.progress_icon:
    displayName: " "
  formula.progress_string:
    displayName: Tasks that are
views:
  - type: table
    name: Таблиця
    groupBy:
      property: formula.progress_string
      direction: DESC
    order:
      - formula.progress_icon
      - file.name
      - formula.priority
      - status
      - blocked by
      - epic
    sort:
      - property: formula.progress_sort
        direction: ASC
      - property: priority
        direction: DESC
      - property: file.basename
        direction: ASC
    summaries: {}
    columnSize:
      formula.progress_icon: -56
      note.status: 125
      note.blocked by: 332

```
<!-- end unchanged block -->

# What to build

Turn the `text convert` command of the private `sbt-mele` project into a self-contained public tool, `excel-to-json`, and make it installable in any GitHub workflow with one step. The tool turns a **Translation spreadsheet** into **Text JSON**; the terms are defined in the glossary. `sbt-mele` itself is not touched: no shared module, no removal of its command, and the code is copied from its working tree (which has uncommitted changes) without history.

**The tool.** A single-purpose program with no subcommands. It takes the spreadsheet through `-i/--input`, writes the JSON to the file named by `-o/--output` or to standard output, and takes the **Gender separator** through `--separator` (` /// ` by default). Its behaviour is that of `text convert`, with one deliberate change: a translation containing the gender separator always yields two items, male then female, either of which may be empty, so an explicitly empty female form is not mistaken for a shared text (`Командир /// ` yields `["Командир", ""]`, and a lone separator yields `["", ""]`). It also gains `--version`, which prints the **Version**. Tests move with the code and pin the new output.

**The release.** The **Version** is the Roman numeral git tag described in [[ADR-0001 - Name a release by a Roman numeral tag|ADR-0001]]; it is the only version of the program and of the Action, and nothing edits a file to cut a release. Pushing a tag builds the program for Linux, Windows and macOS on amd64 and arm64 and creates a draft GitHub release; the maintainer publishes it by hand. The assets are named without the version, so a fixed URL always points at the newest release.

**The Action.** The repository itself is the GitHub Action, with no second repository for it. It installs the newest published release of the program onto the runner's `PATH` on Linux, macOS and Windows, takes no version input, and contains no shell script of its own: it wraps the third-party install action already used by `install-sbt-infinity`, pinned by commit SHA.

**Housekeeping.** The licence is `BlueOak-1.0.0` throughout, replacing the headers of the copied code. The README is a single English sentence. Continuous integration runs the tests on all three operating systems and checks that the Action installs a working program by running `--version` after it.

The publication of a release and the installation of a release with a non-semantic tag by the third-party action are not verified yet. The first real release is therefore preceded by a throwaway release of `I` on a test repository that exercises the whole chain: tag, draft release, publish, the Action, `--version`.

# Acceptance criteria

- [x] `excel-to-json -i <spreadsheet>.xlsx` prints Text JSON to standard output and `-o <file>` writes it to a file instead, creating missing parent directories; invalid input produces an error on standard error and no partial output; the program has no subcommands.
- [x] A translation with the gender separator always yields two items, male then female: `male /// female` yields `["male", "female"]`, `male /// ` yields `["male", ""]`, ` /// female` yields `["", "female"]`, and a lone separator yields `["", ""]`; a translation without it yields one item, the whole text, untrimmed.
- [x] `--separator` replaces the default ` /// `; an empty separator, a translation with more than one separator, an invalid String ID, and a duplicate String ID fail the whole conversion, with the offending row or String ID named.
- [x] Only translated rows contribute an entry, and a spreadsheet with none produces an empty JSON object.
- [x] `excel-to-json --version` prints the Roman numeral of the tag the program was built from, and a build made without a tag does not claim to be a release.
- [ ] The tests of `text convert` are carried over, adapted to the new output, and pass on Linux, macOS and Windows in continuous integration.
- [ ] Pushing a Roman numeral tag builds the program for Linux, Windows and macOS on amd64 and arm64 and creates a draft GitHub release whose assets have no version in their names; the workflow runs goreleaser without semantic-version validation, as ADR-0001 records.
- [ ] A workflow step `uses: <owner>/excel-to-json@<ref>` puts a working `excel-to-json` on the `PATH` of a Linux, a macOS and a Windows runner, installing the newest published release, with no version input and no shell script in `action.yml`; the third-party action it wraps is pinned by commit SHA.
- [ ] Continuous integration installs the program through the Action in the repository's own workflow and checks that `excel-to-json --version` succeeds.
- [x] Every source file carries the `BlueOak-1.0.0` licence header, the licence text is in the repository, and the README is one English sentence.
- [ ] A throwaway release of `I` on a test repository has gone through tag, draft release, manual publish, installation through the Action, and `--version` before the first real release is cut, and any failure found there is fixed or recorded.
