package main

import (
	"errors"
	"fmt"
	"iter"
	"maps"
	"slices"

	"google.golang.org/genai"
)

var (
	expectedArgs = map[string]iter.Seq[string]{
		"edit_file":  maps.Keys(editFileTool.FunctionDeclarations[0].Parameters.Properties),
		"read_file":  maps.Keys(readFileTool.FunctionDeclarations[0].Parameters.Properties),
		"list_files": maps.Keys(listFilesTool.FunctionDeclarations[0].Parameters.Properties),
	}
)

func FunctionCall(call *genai.FunctionCall) *genai.FunctionResponse {
	response := make(map[string]any)
	var err error

	// Check tool and argument names.
	name := call.Name
	received := maps.Keys(call.Args)
	if expected, ok := expectedArgs[name]; !ok {
		message := fmt.Sprintf("tool named '%s' is unknown", name)
		err = errors.New(message)
		response["error"] = message
	} else if want, have := slices.Sorted(expected), slices.Sorted(received); !slices.Equal(want, have) {
		message := fmt.Sprintf("for tool named '%s' want arguments named %v have arguments named %v",
			name, slices.Collect(expected), slices.Collect(received))
		err = errors.New(message)
		response["error"] = message
	} else if call.Name == "read_file" {
		content, err := ReadFile(call.Args["path"].(string))
		if err == nil {
			response["output"] = content
		} else {
			response["error"] = err.Error()
		}
	} else if call.Name == "edit_file" {
		err = EditFile(call.Args["path"].(string), call.Args["old_string"].(string), call.Args["new_string"].(string))
		if err == nil {
			response["output"] = "OK"
		} else {
			response["error"] = err.Error()
		}
	} else if call.Name == "list_files" {
		files, err := ListFiles(call.Args["path"].(string))
		if err == nil {
			response["output"] = files // json encoding of the files list
		} else {
			response["error"] = err.Error()
		}
	} else {
		message := fmt.Sprintf("tool '%s' is unknown", call.Name)
		err = errors.New(message)
		response["error"] = message
	}

	if _, hasOutput := response["output"]; hasOutput {
		greenf("calling '%s' \u2714\n", call.Name)
	} else if _, hasError := response["error"]; hasError {
		redf("calling '%s' \u2718: %v\n", call.Name, err)
	}

	return &genai.FunctionResponse{
		ID:       call.ID,
		Name:     call.Name,
		Response: response,
	}
}
