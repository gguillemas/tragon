# Tragon

[![GoDoc](https://godoc.org/github.com/gguillemas/tragon?status.svg)](https://godoc.org/github.com/gguillemas/tragon)

Tragon is a minimal Go package to build fake SMTP servers that process all incoming data.

The use cases for such servers can be:

- Testing email clients.
- Analyzing email for malware.
- Analyzing email for phishing.
- Harvesting email addresses.
- Collecting email spam.

This package is **not** intended to be used for:

- Building fully working SMTP servers.
- Building interactive mocks of SMTP servers.

## Features

- Always replies positively.
- Customizable SMTP reply messages.
- Customizable connection handling.
- Customizable message handling.
- Customizable errror handling.

## Installation

Execute:

```
go get github.com/gguillemas/tragon
```

## Documentation

Execute:

```
go doc github.com/gguillemas/tragon
```

Or visit [godoc.org](https://godoc.org/github.com/gguillemas/tragon) to read it online.

## Example

An example of a common use case can be found in `cmd/tragon-attachments-example`.

This example shows how Tragon can be used to extract, store and analyze email attachments.
