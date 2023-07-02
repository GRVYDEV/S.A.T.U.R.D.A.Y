package openai

import (
	"context"
	"errors"
	"fmt"

	logr "github.com/GRVYDEV/S.A.T.U.R.D.A.Y/log"
	"github.com/GRVYDEV/S.A.T.U.R.D.A.Y/ttt/engine"

	openai "github.com/sashabaranov/go-openai"
)

var _ engine.Generator = (*OpenAIBackend)(nil)
var Logger = logr.New()

const PREFIX = `
Your name is Saturday.

Saturday is a conversational, vocal, artificial intelligence assistant.

Saturday's job is to converse with humans to help them accomplish goals.

Saturday is able to help with a wide variety of tasks from answering questions to assisting the human with creative writing.

Overall Saturday is a powerful system that can help humans with a wide range of tasks and provide valuable insights as well as taking actions for the human.
`

const SUFFIX = `
DO NOT start your response with "Assistant:"

Given the following prompt respond as a friendly assistant:

PROMPT: %s
`

type OpenAIBackend struct {
	client *openai.Client
}

func New(token string) (*OpenAIBackend, error) {
	if token == "" {
		return nil, errors.New("cannot create OpenAIBackend without a token")
	}

	return &OpenAIBackend{
		client: openai.NewClient(token),
	}, nil
}

func (o *OpenAIBackend) Generate(prompt string) (engine.TextChunk, error) {
	var (
		chunk engine.TextChunk
		err   error
	)
	resp, err := o.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: PREFIX,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: fmt.Sprintf(SUFFIX, prompt),
				},
			},
		},
	)

	if err != nil {
		Logger.Error(err, "ChatCompletion error")
		return chunk, err
	}

	if len(resp.Choices) == 0 {
		return chunk, errors.New("openai returned empty choices")
	}

	return engine.TextChunk{
		Text: resp.Choices[0].Message.Content,
	}, nil

}
