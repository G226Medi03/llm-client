# llmclient

Go client library for text LLMs with a small, provider-agnostic interface.

Supported providers:

- **DeepSeek** (OpenAI-compatible chat completions, with OCR / image input)
- **Ollama** (local `/api/chat`)
- **Gemini** (Gemini 4 Argon thinking model)

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

### Gemini

```go
client, err := gemini.New(gemini.Config{
    APIKey:        os.Getenv("GEMINI_API_KEY"),
    Model:         gemini.ModelArgon, // gemini-4-argon
    ThinkingLevel: gemini.ThinkingHigh,
})
```

### Factory

```go
client, err := llmclient.New(llmclient.Config{
    Provider:      llmclient.ProviderGemini,
    Model:         gemini.ModelArgon,
    APIKey:        os.Getenv("GEMINI_API_KEY"),
    ThinkingLevel: "high",
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
