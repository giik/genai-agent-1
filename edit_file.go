package main

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"google.golang.org/genai"
)

const editFileDescription = `Edit the contents of the file at the given ` +
	`relative 'path' argument by replacing instances of 'old_string' with 'new_string'. ` +
	`'old_string' and 'new_string' muust be different from each other. If the file specified ` +
	`in 'patyh' does not exist it will be created.`

var editFileTool = &genai.Tool{
	FunctionDeclarations: []*genai.FunctionDeclaration{
		{
			Description: editFileDescription,
			Name:        "edit_file",
			Parameters: &genai.Schema{
				Type: genai.TypeObject,
				Properties: map[string]*genai.Schema{
					"path": {
						Type: genai.TypeString,
					},
					"old_string": {
						Type: genai.TypeString,
					},
					"new_string": {
						Type: genai.TypeString,
					},
				},
			},
		},
	},
}

func CreateNewFile(filePath, content string) error {
	dir := path.Dir(filePath)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create directory: %w", err)
		}
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return fmt.Errorf("failed to create %s: %w", filePath, err)
	}

	return nil
}

func EditFile(path, oldStr, newStr string) error {
	if path == "" {
		return errors.New("Invalid argument: file path is empty.")
	}

	oldContent, err := ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) && oldStr == "" {
			return CreateNewFile(path, newStr)
		}
		return err
	}

	if oldStr == "" {
		return errors.New("Invalid argument: 'old_string' is empty.")
	}

	newContent := strings.Replace(oldContent, oldStr, newStr, -1)
	if oldContent == newContent && oldStr != "" {
		return fmt.Errorf("'old_string' not found in %s", path)
	}

	err = os.WriteFile(path, []byte(newContent), 0644)
	if err != nil {
		return err
	}
	return nil
}
