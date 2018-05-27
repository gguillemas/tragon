# Tragon

Tragon is a minimal Go package to build fake SMTP servers that process all incoming data.

The use cases for such servers can be:

- Testing email clients.
- Analyzing email for malware.
- Analyzing email for phishing.
- Harvesting email addresses.
- Collecting email spam.

This package is **not** intented to be used for:

- Building fully working SMTP servers.
- Building interactive mocks of SMTP servers.

## Installation

```go get github.com/gguillemas/tragon```

## Example

An example of a common use case can be found in `cmd/tragon-attachments-example`.

This example shows how Tragon can be used to extract, store and analyze email attachments.
