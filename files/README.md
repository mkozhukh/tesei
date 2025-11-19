# Files

The `files` package provides jobs for file system operations and file content processing.

## Types

### `TextFile`
Represents a text file with name, folder path, and content.

```go
type TextFile struct {
    Name    string
    Folder  string
    Content string
}
```

## Jobs

### `ListDir`
Lists files in a directory. Supports recursion, filtering, and limits.

```go
files.ListDir{
    Path: "./data",
    Ext:  ".txt",
    Nested: true,
}
```

### `ReadFile`
Reads the content of files passed in the pipeline.

```go
files.ReadFile{}
```

### `WriteFile`
Writes content to files. Can change destination folder.

```go
files.WriteFile{
    Folder: "./output",
    Overwrite: true,
}
```

### `PrintContent`
Prints the ID and content of the file to stdout.

```go
files.PrintContent{}
```

### `HashContent`
Calculates a hash of the content and stores it in metadata.

```go
files.HashContent{
    Key: "hash", // Metadata key
    Size: 8,     // Hash length
}
```

### `RenameFile`
Renames the file (in memory, for subsequent write). Supports template replacement from metadata.

```go
files.RenameFile{
    Suffix: "_{{hash}}",
    Ext: ".md",
}
```

### `Replace`
Replaces strings in content using a map. Supports template replacement in values.

```go
files.Replace{
    Matches: map[string]string{
        "foo": "bar",
        "baz": "{{hash}}",
    },
}
```

### `Filter`
Filters files based on a custom function.

```go
files.Filter{
    Match: func(msg *tesei.Message[files.TextFile]) bool {
        return len(msg.Data.Content) > 0
    },
}
```

### `Split`
Splits a file into multiple chunks based on a user-defined rule. Adds metadata (`split_id`, `split_index`, `split_total`) for later merging.

```go
files.Split{
    By: func(text string) []string {
        return strings.Split(text, "\n\n") // Split by paragraph
    },
}
```

### `Merge`
Merges chunks back into a single file. Expects `split_id`, `split_index`, and `split_total` metadata.

```go
files.Merge{
    Glue: "\n\n", // Join with double newline
    // Or use a custom function:
    // By: func(chunks []string) string { ... },
}
```

### `Clone`
Generates multiple messages from a single input message using a custom handler. Useful for creating variants of a file.

```go
files.Clone{
    By: func(msg *tesei.Message[files.TextFile]) []*tesei.Message[files.TextFile] {
        m1 := msg.Clone()
        m1.Data.Name += ".bak"
        return []*tesei.Message[files.TextFile]{msg, m1}
    },
}
```
