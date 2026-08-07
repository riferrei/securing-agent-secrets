// Package agent runs the reasoning loop: the model picks a customer id, the Go
// code owns the Redis call.
package agent

import (
	"context"
	"fmt"
	"log"

	"github.com/riferrei/securing-agent-secrets-1password/internal/llm"
	"github.com/riferrei/securing-agent-secrets-1password/internal/redisstore"
)

type TurnResult struct {
	UserMessage      string
	AssistantMessage string
	ToolUsed         string
	RedisCommand     string
	RedisResult      string
}

type Agent struct {
	llm      *llm.Client
	store    *redisstore.Store
	model    string
	maxIters int
}

func New(client *llm.Client, store *redisstore.Store, model string, maxIters int) *Agent {
	if maxIters <= 0 {
		maxIters = 5
	}
	return &Agent{llm: client, store: store, model: model, maxIters: maxIters}
}

var getCustomerTool = llm.Tool{
	Type: "function",
	Function: llm.ToolFunction{
		Name:        "get_customer",
		Description: "Look up a customer record by id. Ids are numeric, e.g. 1 for customer 0001.",
		Parameters: llm.ToolParameters{
			Type: "object",
			Properties: map[string]llm.Property{
				"customer_id": {Type: "string", Description: "The customer id, e.g. 1 or 0001"},
			},
			Required: []string{"customer_id"},
		},
	},
}

const systemPrompt = `You are a data assistant for a customer support team.
When the user asks about a customer, call the get_customer tool with the id.
Answer using the tool results. Keep answers short and factual.`

func (a *Agent) Ask(ctx context.Context, question string) (TurnResult, error) {
	res, _, err := a.Respond(ctx, nil, question)
	return res, err
}

// Respond continues a conversation; the caller persists the returned history
// for the next turn.
func (a *Agent) Respond(ctx context.Context, history []llm.Message, question string) (TurnResult, []llm.Message, error) {
	res := TurnResult{UserMessage: question, ToolUsed: "none"}

	messages := history
	if len(messages) == 0 {
		messages = []llm.Message{{Role: "system", Content: systemPrompt}}
	}
	messages = append(messages, llm.Message{Role: "user", Content: question})
	tools := []llm.Tool{getCustomerTool}

	for i := 0; i < a.maxIters; i++ {
		reply, err := a.llm.Chat(ctx, a.model, messages, tools)
		if err != nil {
			return res, messages, fmt.Errorf("model call failed: %w", err)
		}
		messages = append(messages, *reply)

		if len(reply.ToolCalls) == 0 {
			res.AssistantMessage = reply.Content
			return res, messages, nil
		}

		for _, tc := range reply.ToolCalls {
			content := a.runToolCall(ctx, tc, &res)
			messages = append(messages, llm.Message{
				Role:     "tool",
				ToolName: tc.Function.Name,
				Content:  content,
			})
		}
	}

	// Out of tool rounds: one more call without tools to force a final answer.
	reply, err := a.llm.Chat(ctx, a.model, messages, nil)
	if err != nil {
		return res, messages, fmt.Errorf("model call failed: %w", err)
	}
	messages = append(messages, *reply)
	res.AssistantMessage = reply.Content
	return res, messages, nil
}

func (a *Agent) runToolCall(ctx context.Context, tc llm.ToolCall, res *TurnResult) string {
	switch tc.Function.Name {
	case "get_customer":
		return a.runGetCustomer(ctx, tc.Function.Arguments, res)
	default:
		return `{"error":"unknown tool"}`
	}
}

func (a *Agent) runGetCustomer(ctx context.Context, args map[string]any, res *TurnResult) string {
	rawID := strArg(args, "customer_id")
	if rawID == "" {
		// Some models pass the id as a number.
		if n, ok := args["customer_id"].(float64); ok {
			rawID = fmt.Sprintf("%d", int(n))
		}
	}

	res.ToolUsed = "get_customer"
	res.RedisCommand = "HGETALL " + redisstore.Key(rawID)
	log.Printf("agent: tool call get_customer(%q) -> %s", rawID, res.RedisCommand)

	raw, err := a.store.GetCustomerRaw(ctx, rawID)
	if err != nil {
		msg := `{"error":"customer not found"}`
		res.RedisResult = msg
		return msg
	}
	res.RedisResult = raw
	return raw
}

func strArg(args map[string]any, key string) string {
	if v, ok := args[key].(string); ok {
		return v
	}
	return ""
}
