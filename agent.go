package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"

	"google.golang.org/genai"
)

type Agent struct {
	Client    *genai.Client
	Model     string
	GetPrompt func() (string, bool)
	LogFile   *os.File
}

func NewAgent(client *genai.Client, modelName string, getPrompt func() (string, bool), logFile *os.File) *Agent {
	return &Agent{
		Client:    client,
		Model:     modelName,
		GetPrompt: getPrompt,
		LogFile:   logFile, // nil if logging disabled/not possible
	}
}

type AgentState uint

const (
	StateUndefined AgentState = iota
	Error
	ProcessResponse
	SendMessage
	UserInput
)

// State machine approach makes it a bit cleaner.
func (a *Agent) Run(ctx context.Context, chat *genai.Chat) error {
	var state AgentState = UserInput
	var err error
	var conversation []genai.Part
	var response *genai.GenerateContentResponse

	for {
		switch state {

		case Error:
			return err

		case UserInput:
			bluef("[user]: ")
			userInput, ok := a.GetPrompt()
			if !ok {
				err = errors.New("error getting user input")
				state = Error
			} else {
				AppendToLog(a.LogFile, "user", userInput)
				conversation = []genai.Part{genai.Part{Text: userInput}}
				state = SendMessage
			}

		case SendMessage:
			response, err = chat.SendMessage(ctx, conversation...)
			if err != nil {
				state = Error
			} else {
				state = ProcessResponse
			}

		case ProcessResponse:
			numCandidates := len(response.Candidates)
			if numCandidates == 0 {
				yellowln("received zero candidates from model")
				state = UserInput
				continue
			}
			if numCandidates != 1 {
				redf("asked for 1 but received %d candidates; all but first will be ignored\n", numCandidates)
			}

			content := response.Candidates[0].Content
			numParts := len(content.Parts)
			if numParts == 0 {
				yellowln("first response candidates has no parts, back to user")
				state = UserInput
				continue
			}

			conversation = nil
			var callParts, textParts []string // descriptions of the text and functioncall responses, respectively
			for i, part := range content.Parts {
				if len(part.Text) != 0 {
					// Handle text response.
					textParts = append(textParts, part.Text)
					yellowf("[%s][text %d bytes][part %d of %d] ", content.Role, len(part.Text), i+1, numParts)
					whiteln(part.Text)
				} else if part.FunctionCall != nil {
					// Handle request for function call.
					callParts = append(callParts, fmt.Sprintf("call '%s' (args omitted)", part.FunctionCall.Name))
					magentaf("calling '%s' with args {\n", part.FunctionCall.Name)
					for _, arg := range DescribeMapEntries(part.FunctionCall.Args) {
						whitef("'%s':(%s)\n", arg.Key, arg.Description)
					}
					magentaln("}")
					conversation = append(conversation, genai.Part{FunctionResponse: FunctionCall(part.FunctionCall)})
				} else {
					yellowln("only Text and FunctionCall are supported!")
				}
			}
			// Textual descriptions to be written to the markdown log.
			calls := fmt.Sprintf("%d function calls", len(callParts))
			texts := strings.Join(textParts, "\n\n")
			if len(textParts) == 0 {
				if len(callParts) == 0 {
					AppendToLog(a.LogFile, "model", "No text or function call responses received")
					state = UserInput
				} else {
					AppendToLog(a.LogFile, "model tool", calls)
					state = SendMessage
				}
			} else {
				if len(callParts) == 0 {
					AppendToLog(a.LogFile, "model text", texts)
					state = UserInput
				} else {
					AppendToLog(a.LogFile, "model text + tool", fmt.Sprintf("%s\n\n**function calls**", texts, calls))
					state = SendMessage
				}
			}
		}
	}
}

// EntryInfo holds the key and a description of the value's type and content.
type EntryInfo struct {
	Key         string
	Description string
}

// DescribeMapEntries takes a map[string]any and returns a slice of EntryInfo
// describing each entry's key and value.
func DescribeMapEntries(data map[string]any) []EntryInfo {
	var result []EntryInfo

	for key, value := range data {
		description := ""
		switch v := value.(type) {
		case nil:
			description = "nil value"
		case string:
			description = fmt.Sprintf("string of length %d", len(v))
		case int, int8, int16, int32, int64:
			description = fmt.Sprintf("int %v", v)
		case uint, uint8, uint16, uint32, uint64, uintptr:
			description = fmt.Sprintf("uint %v", v)
		case float32, float64:
			description = fmt.Sprintf("float %v", v)
		case bool:
			description = fmt.Sprintf("boolean equal to %v", v)
		// Handle slices and maps using reflection for more detail
		default:
			val := reflect.ValueOf(value)
			switch val.Kind() {
			case reflect.Slice, reflect.Array:
				elemType := val.Type().Elem().String()
				description = fmt.Sprintf("array of %s of length %d", elemType, val.Len())
			case reflect.Map:
				keyType := val.Type().Key().String()
				elemType := val.Type().Elem().String()
				description = fmt.Sprintf("map of %s to %s of length %d", keyType, elemType, val.Len())
			case reflect.Struct:
				description = fmt.Sprintf("struct of type %s", val.Type().String())
			case reflect.Ptr:
				description = fmt.Sprintf("pointer to %s", val.Type().Elem().String())
			case reflect.Func:
				description = fmt.Sprintf("function of type %s", val.Type().String())
			case reflect.Chan:
				description = fmt.Sprintf("channel of type %s", val.Type().String())
			case reflect.Interface:
				description = fmt.Sprintf("an interface") // could add more detail if interface is not nil
			default:
				description = fmt.Sprintf("value of type %T", value)
			}
		}

		result = append(result, EntryInfo{
			Key:         key,
			Description: description,
		})
	}

	return result
}
