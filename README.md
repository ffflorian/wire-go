# wire-go [![Build Status](https://github.com/ffflorian/wire-go/workflows/Build/badge.svg)](https://github.com/ffflorian/wire-go/actions/)

A [Wire](https://github.com/wireapp) CLI.

## Installation

Run `go get github.com/ffflorian/wire-go`.

## Local usage

```
wire-go
```

```
A Wire CLI.

Usage:
  wire-go [options] [directory]

Options:
  --client-id, -i      specify the client's ID (e.g. for setting its label)
  --label, -l          specify the client's new label
  --backend, -b        specify the Wire backend URL (default: "staging-nginz-https.zinfra.io")
  --email, -e          specify your Wire email address
  --password, -p       specify your Wire password
  --version, -v        output the version number
  --help, -h           display this help

Commands:
  delete-all-clients
  set-client-label
  get-all-clients
```
