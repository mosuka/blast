# Bleve Language Detection

## detect_lang token filter

A bleve token filter which passes the text of each token and passes it to the cld2 library.  The library determines what it thinks the language most likely is.  The ISO-639 language code replaces the token term.

## detect_lang analyzer

An analyzer configured to treat the entire input as a single token, the input is lower-cased, and the passed to the detect_lang token filter.

When you configure a field to use this analyzer, the effect will be to index a single token with the value being the detected ISO-639 language code.

### Dependency

This language uses a [Go wrapper](https://github.com/blevesearch/cld2) around [CLD2](https://github.com/CLD2Owners/cld2).  The wrapper includes the C++ source directly so previous complications around static vs dynamic linking have gone away.
