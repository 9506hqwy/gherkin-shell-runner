# Gherkin Shell Runner

This repository provides the TUI test runner that run tests written in Gherkin `.feature` files.

## Features

- Run shell commands from Gherkin scenarios
- Pass arguments with `arg` steps
- Provide stdin content with `stdin` step
- Set working directory with `workspace` step
- Set environment variables with `env` step
- Assert exit status and command output
- Filter scenarios by tags

## Build

```sh
go build -o bin/gherkin-shell-runner ./cmd/gherkin-shell-runner
```

## Usage

```sh
./bin/gherkin-shell-runner [flags] [feature_files_or_directories]
```

Example:

```sh
./bin/gherkin-shell-runner examples/features
```

## Supported Step Definitions

### Arrange

- `Given command <command>`
- `Given workspace <path>`
- `Given env <name> <value>`
- `Given arg <argument>`
- `Given stdin <text>`
- `Given stdin` followed by a doc string block
- `Given wait <milli second>`
- `Given timeout <milli second>`
- `Given size <width> <height>`
- `Given encoding output <encoding>`
- `Given use temp workspace`
- `Given set <name> <value>`

### Act

- `When exec`

### Assert

- `Then status eq <code>`
- `Then status not eq <code>`
- `Then output is empty`
- `Then output is not empty`
- `Then output eq <text>`
- `Then output eq` followed by a doc string block
- `Then output not eq <text>`
- `Then output not eq` followed by a doc string block
- `Then output regex <text>`
- `Then output regex` followed by a doc string block
- `Then output not regex <text>`
- `Then output not regex` followed by a doc string block

## Example Feature

see [examples](./examples/features/).

## Help

```sh
$ ./bin/gherkin-shell-runner -h
Gherkin Shell Runner

Usage:
  gherkin-shell-runner [flags]

Flags:
      --concurrency int    Run scenario concurrency. (default 1)
      --format string      Report format. (default "progress")
  -h, --help               help for gherkin-shell-runner
      --no-colors          Disable ansi color.
      --random int         Randamize scenario order. (default -1)
      --show-steps         Show avaiblae step definitions.
      --stop-on-failture   Stop on first failed scenario.
      --tags string        Filter scenario. (default "~@ignore")
  -v, --version            version for gherkin-shell-runner
```

## TODO

- Wait for complete output correctly.
- More information in HTML report.
- Input encoding.
- Timeout handling.
- Newline handling.
- File operation
- Expression
- DataTable
