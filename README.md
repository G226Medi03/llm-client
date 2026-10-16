# llmclient

Go client library for text LLMs with a small, provider-agnostic interface.

Supported providers:

- **DeepSeek** (OpenAI-compatible chat completions, with OCR / image input)
- **Ollama** (local `/api/chat`)

## Install

```sh
go get github.com/G226Medi03/llm-client
```

## Interfaces

```go
type Client interface {
    Model() string
    Generate(ctx context.Context, prompt string) (string, error)
    GenerateStream(ctx context.Context, prompt string) (<-chan StreamResult, error)
    ResetContext()
    Clone() Client
}

// Implemented by clients that accept images (e.g. DeepSeek).
type OcrClient interface {
    AskOCR(ctx context.Context, imagePath string) string
    GenerateWithImage(ctx context.Context, prompt string, imagePath string) (string, error)
}
```

## Usage

### DeepSeek

```go
client, err := deepseek.New(deepseek.Config{
    APIKey: os.Getenv("DEEPSEEK_API_KEY"),
    Model:  deepseek.ModelV4Flash,
})
if err != nil {
    log.Fatal(err)
}

text, err := client.Generate(ctx, "안녕하세요")
```

### Ollama

```go
client, err := ollama.New(ollama.Config{
    BaseURL: "http://localhost:11434",
    Model:   "gemma4:e4b",
})
```

### Factory

```go
client, err := llmclient.New(llmclient.Config{
    Provider: llmclient.ProviderDeepSeek,
    Model:    deepseek.ModelV4Flash,
    APIKey:   os.Getenv("DEEPSEEK_API_KEY"),
})
```

### Options

```go
client, err := deepseek.New(deepseek.Config{
    APIKey:   os.Getenv("DEEPSEEK_API_KEY"),
    JSONMode: true,          // request JSON output
    Logger:   slog.Default(), // optional
})
```

## License

MIT
