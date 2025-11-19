# LLM

The `llm` package provides jobs for integrating with Large Language Models (LLMs). It uses `github.com/mkozhukh/echo` for LLM clients and `github.com/mkozhukh/echo-templates` for prompt management.

## Configuration

Before using LLM jobs, you must configure the model and optionally the templates source.

```go
llm.SetModel("openai/gpt-4o")
llm.SetAPIKey(os.Getenv("OPENAI_API_KEY")) // Optional if env var is set
llm.SetTemplatesPath("./templates")        // Or use SetTemplatesSource
```

## Jobs

### `CompleteContent`
Sends the file content to the LLM and replaces the content with the response.

```go
llm.CompleteContent{
    Prompt: "Summarize this text", // Optional system prompt
}
```

### `CompleteTemplateString`
Uses an inline template string to generate content.

```go
llm.CompleteTemplateString{
    Template: "@system: You are a helper.\n@user: {{user_query}}",
}
```

### `CompleteTemplate`
Uses a named template from the configured templates source.

```go
llm.CompleteTemplate{
    Template: "summarize",
    Vars: map[string]any{
        "extra_context": "some value",
    },
}
```
