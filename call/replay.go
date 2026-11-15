package call

import (
	"context"
	"errors"
)

// ReplayCall is one stored LLM call used to replay a prior run.
type ReplayCall struct {
	Stage  string
	Prompt string
	Output string
}

// ReplayOptions configures a replay client.
type ReplayOptions struct {
	// Calls are the stored calls indexed by prompt and stage.
	Calls []ReplayCall
	// Live is used for stages enabled in LiveStages. It may be nil for
	// replay-only runs.
	Live *Client
	// LiveStages marks stages that must call Live instead of using stored
	// output.
	LiveStages map[string]bool
	// StageOf infers the stage from a prompt. When nil, only prompt matching
	// is available.
	StageOf func(prompt string) string
}

// NewReplay creates a Client backed by stored LLM calls.
//
// Stored calls are matched by exact prompt first, then by stage. Live stages
// call the live client's generate function directly so a replay call does not
// acquire a second concurrency slot.
func NewReplay(opts ReplayOptions) *Client {
	outputs := make(map[string]string, len(opts.Calls))
	byStage := make(map[string]string, len(opts.Calls))
	for _, call := range opts.Calls {
		if call.Prompt != "" {
			outputs[call.Prompt] = call.Output
		}
		if call.Stage != "" {
			byStage[call.Stage] = call.Output
		}
	}

	model := "replay"
	if opts.Live != nil {
		model = opts.Live.Model()
	}

	return NewFunc(model, func(ctx context.Context, prompt string) (string, error) {
		stage := ""
		if opts.StageOf != nil {
			stage = opts.StageOf(prompt)
		}
		if stage != "" && opts.LiveStages[stage] {
			if opts.Live == nil || opts.Live.generate == nil {
				return "", errors.New("replay live client is required")
			}
			return opts.Live.generate(ctx, prompt)
		}
		if output, ok := outputs[prompt]; ok {
			return output, nil
		}
		if stage != "" {
			if output, ok := byStage[stage]; ok {
				return output, nil
			}
		}
		return "", errors.New("replay raw output 이 없습니다")
	})
}
