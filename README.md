# llmclient

Go client library for text LLMs with a small, provider-agnostic interface.

All public API lives in the root `llmclient` package. Provider implementations
are internal and cannot be imported by other modules.

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
type LlmResponse string

type LlmClient interface {
    Model() string
    Generate(ctx context.Context, prompt string) (LlmResponse, error)
    GenerateStream(ctx context.Context, prompt string) (<-chan StreamResult, error)
    ResetContext()
    Clone() LlmClient
}

type OcrClient interface {
    AskOCR(ctx context.Context, imagePath string) string
    GenerateWithImage(ctx context.Context, prompt string, imagePath string) (string, error)
}

// ImageClient combines LlmClient and OcrClient.
type ImageClient interface {
    LlmClient
    OcrClient
}
```

## Usage

### Unified factory

```go
client, err := llmclient.New(llmclient.Config{
    Provider: llmclient.ProviderOpenCode,
    APIKey:   os.Getenv("OPENCODE_API_KEY"),
    BaseURL:  "https://opencode.ai/inference/go",
    Model:    "deepseek-v4.1-flash",
    API:      llmclient.OpenCodeAPIOpenAI,
})
if err != nil {
    log.Fatal(err)
}

response, err := client.Generate(ctx, "안녕하세요")
fmt.Println(string(response))
```

### DeepSeek

`NewDeepSeek` returns an `ImageClient`, so OCR methods are available directly.

```go
client, err := llmclient.NewDeepSeek(llmclient.DeepSeekConfig{
    APIKey: os.Getenv("DEEPSEEK_API_KEY"),
    Model:  llmclient.ModelDeepSeekV4Flash,
})
if err != nil {
    log.Fatal(err)
}

response, err := client.Generate(ctx, "안녕하세요")
fmt.Println(string(response))
```

### Ollama

```go
client, err := llmclient.NewOllama(llmclient.OllamaConfig{
    BaseURL: llmclient.DefaultOllamaBaseURL,
    Model:   llmclient.ModelOllamaGemma4E4b,
})
```

### Gemini

```go
client, err := llmclient.NewGemini(llmclient.GeminiConfig{
    APIKey:        os.Getenv("GEMINI_API_KEY"),
    Model:         llmclient.ModelGeminiArgon,
    ThinkingLevel: llmclient.GeminiThinkingHigh,
})
```

### OpenCode

```go
client, err := llmclient.NewOpenCode(llmclient.OpenCodeConfig{
    APIKey:  os.Getenv("OPENCODE_API_KEY"),
    BaseURL: "https://opencode.ai/inference/go",
    Model:   "deepseek-v4.1-flash",
    API:     llmclient.OpenCodeAPIOpenAI,
})
```

Solver / Responses API:

```go
solver, err := llmclient.NewOpenCode(llmclient.OpenCodeConfig{
    APIKey:  os.Getenv("OPENCODE_CONSOLE_API_KEY"),
    BaseURL: "https://opencode.ai/inference",
    Model:   "gpt-6.1-sol",
    API:     llmclient.OpenCodeAPIResponses,
})
```

### OCR

```go
client, err := llmclient.NewDeepSeek(llmclient.DeepSeekConfig{
    APIKey: os.Getenv("DEEPSEEK_API_KEY"),
})
if err != nil {
    log.Fatal(err)
}

text := client.AskOCR(ctx, "./image.png")
answer, err := client.GenerateWithImage(ctx, "이미지를 설명해줘", "./image.png")
```

### Isolated single call

`GenerateIsolated` clones the client before each call so parallel callers do
not share conversation state.

```go
answer, err := llmclient.GenerateIsolated(ctx, client, prompt, 30*time.Second)
```

### Options

```go
client, err := llmclient.NewDeepSeek(llmclient.DeepSeekConfig{
    APIKey: os.Getenv("DEEPSEEK_API_KEY"),
    Options: llmclient.Options{
        JSONMode: true,
        Logger:   slog.Default(),
    },
})
```

## Replay

`call.NewReplay` builds a `*call.Client` backed by stored calls. Prompt matches
take priority; stage matches are used as a fallback. Stages marked in
`LiveStages` call the live client directly, without nested concurrency slots.

```go
replay := call.NewReplay(call.ReplayOptions{
    Calls: []call.ReplayCall{
        {Stage: "status_extract", Prompt: "prompt", Output: "stored output"},
    },
    Live:       liveClient,
    LiveStages: map[string]bool{"relation_extract": true},
    StageOf:    func(prompt string) string { return "status_extract" },
})
```

`Live` may be nil for replay-only runs. `StageOf` may be nil when only exact
prompt matching is needed.

## Migration from v0.1.6

The provider packages are now internal. Replace imports such as:

```go
import "github.com/G226Medi03/llm-client/deepseek"
```

with the root facade:

```go
import "github.com/G226Medi03/llm-client"
```

Use `llmclient.New`, `llmclient.NewDeepSeek`, `llmclient.NewGemini`,
`llmclient.NewOllama`, or `llmclient.NewOpenCode` instead of provider
constructors. Use `llmclient.GenerateIsolated` instead of
`llm-client/call.Generate`.

## License

MIT
