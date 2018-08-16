## this fork

This is a fork of [this original repo](https://github.com/rainycape/cld2)  Basically just added a few test cases before integrating with it.

# cld2
--
    import "cld2"

Package cld2 implements language detection using the Compact Language Detector.

This package includes the relevant sources from the cld2 project, so it doesn't
require any external dependencies. For more information about CLD2, see
https://code.google.com/p/cld2/.

## Usage

#### func  Detect

```go
func Detect(text string) string
```
Detect returns the language code for detected language in the given text.
