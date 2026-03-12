# NOTE: This file controls license-header generation for the Apollo provider scaffold.
schema_version = 1

project {
  license        = "MPL-2.0"
  copyright_year = 2021
  copyright_holder = "Apollo Graph, Inc."

  header_ignore = [
    # internal catalog metadata (prose)
    "META.d/**/*.yaml",

    # rendered docs and custom tfplugindocs templates (prose)
    "docs/**",
    "templates/**",
    "dist/**",

    # examples used within documentation (prose)
    "examples/**",

    # workflow and issue template configuration
    ".github/workflows/*.yml",

    # GitHub issue template configuration
    ".github/ISSUE_TEMPLATE/*.yml",

    # golangci-lint tooling configuration
    ".golangci.yml",

    # GoReleaser tooling configuration
    ".goreleaser.yml",

    # top-level prose
    "README.md",
    "CHANGELOG.md",
  ]
}
