# devske/localizer

High-performance, thread-safe internationalization (i18n) and localization engine for Go microservices.

## Features

- **100% Thread-Safe**: Concurrent read/write with `sync.RWMutex` ensuring zero data races in high-concurrency HTTP servers.
- **Fast Interpolation**: Token replacement supporting both `{param}` and `{{.param}}` without expensive runtime template compilation.
- **Hierarchical Asset Loading**: Scans directory structures supporting YAML (`.yaml`, `.yml`) and JSON (`.json`) files.
- **Drop-in Compatibility**: Fully implements the legacy interface while introducing modern, request-safe translation methods (`T()`).

## Installation

```bash
go get github.com/devsinke/localizer
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/devsinke/localizer"
)

func main() {
    // Initialize with path to translation directory
    loc := localizer.New("./localizations_src")
    loc.SetDefaultLocale("en", "en")
    loc.SetupLocalization()

    // 1. Thread-safe direct translation for a specific locale
    msg := loc.T("en", "response.welcome", map[string]string{"user": "Alex"})
    fmt.Println(msg)

    // 2. Standard compatibility call
    loc.SetLocale("sw")
    swMsg := loc.Translate("response.welcome", map[string]string{"user": "Alex"})
    fmt.Println(swMsg)
}
```

## Directory Structure Example

```
localizations_src/
├── en/
│   └── response.yaml
├── fr/
│   └── response.yaml
└── sw/
    └── response.yaml
```

## License

MIT License. Copyright (c) 2026 Devs-In-Kenya (devske).
