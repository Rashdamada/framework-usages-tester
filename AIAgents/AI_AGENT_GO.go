package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "time"

    openai "github.com/sashabaranov/go-openai"
)

func main() {
    ctx := context.Background()
    client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

    // 1. Ask the model to call get_current_time
    resp, err := client.Chat.Completion(ctx, openai.ChatCompletionRequest{
        Model: openai.GPT4o,
        Messages: []openai.ChatCompletionMessage{
            {
                Role:    openai.ChatMessageRoleUser,
                Content: "Please tell me the current UTC time.",
            },
        },
        Functions: []openai.FunctionDefinition{
            {
                Name:        "get_current_time",
                Description: "Returns the current UTC time in ISO8601 format.",
                Parameters: map[string]interface{}{
                    "type":       "object",
                    "properties": map[string]interface{}{},
                },
            },
        },
        FunctionCall: "auto", // Let model choose to call the function
    })

    if err != nil {
        log.Fatalf("Chat completion error: %v", err)
    }

    msg := resp.Choices[0].Message

    var output string
    if msg.FunctionCall != nil && msg.FunctionCall.Name == "get_current_time" {
        // 2. Execute the tool
        now := time.Now().UTC().Format(time.RFC3339)
        output = now

        // 3. Send the function response back to the model
        resp2, err := client.Chat.Completion(ctx, openai.ChatCompletionRequest{
            Model: openai.GPT4o,
            Messages: []openai.ChatCompletionMessage{
                {Role: openai.ChatMessageRoleUser, Content: "Please tell me the current UTC time."},
                {
                    Role:    openai.ChatMessageRoleFunction,
                    Name:    "get_current_time",
                    Content: output,
                },
            },
        })
        if err != nil {
            log.Fatalf("Function response error: %v", err)
        }
        fmt.Println(resp2.Choices[0].Message.Content)
    } else {
        // 4. Orchestrate model response if no function was called
        fmt.Println(msg.Content)
    }
}
