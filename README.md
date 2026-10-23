# llmclient

Go client library for text LLMs with a small, provider-agnostic interface.

Supported providers:

- **DeepSeek** (OpenAI-compatible chat completions, with OCR / image input)
- **Ollama** (local `/api/chat`)
- **Gemini** (Gemini 4 Argon thinking model)
- **OpenCode** (OpenCode inference gateway: OpenAI chat/completions, Responses, OCR)

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

### OpenCode

```go
client, err := opencode.New(opencode.Config{
    APIKey:  os.Getenv("OPENCODE_API_KEY"),
    BaseURL: "https://opencode.ai/inference/go",
    Model:   "deepseek-v4.1-flash",
    API:     opencode.APIOpenAI,
})
```

Solver / Responses API:

```go
solver, err := opencode.New(opencode.Config{
    APIKey:  os.Getenv("OPENCODE_CONSOLE_API_KEY"),
    BaseURL: "https://opencode.ai/inference",
    Model:   "gpt-6.1-sol",
    API:     opencode.APIResponses,
})
```

### Factory

```go
client, err := llmclient.New(llmclient.Config{
    Provider:      llmclient.ProviderOpenCode,
    Model:         "gpt-6.1-sol",
    APIKey:        os.Getenv("OPENCODE_CONSOLE_API_KEY"),
    BaseURL:       "https://opencode.ai/inference",
    API:           "responses",
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
