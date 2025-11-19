# Text

The `text` package provides jobs for text processing and cleaning, particularly useful for Markdown and LLM outputs.

## Jobs

### `Markdown`
Provides utilities for processing Markdown files.

- `EscapeTagsInContent`: Escapes HTML-like tags in content to prevent them from being rendered as HTML (except in code blocks).
- `LowerCaseLinks`: Converts internal Markdown links to lowercase.

```go
text.Markdown{
    EscapeTagsInContent: true,
    LowerCaseLinks:      true,
}
```

### `CleanAfterLLM`
Cleans up common artifacts from LLM generation, such as replacing special arrow characters with standard `->`, normalizing dashes, and removing zero-width characters.

```go
text.CleanAfterLLM{}
```
