package main

// Expects that the following environment variables are set:
//
// GOOGLE_API_KEY contains the API KEY for the LLM (without a key a conversation
//                with the agent cannot be established)
//
// CODE_AGENT_LOGDIR contains a directory where session logs will be saved (if
//                   not set /tmp will be used)

import (
	"bufio"
	"context"
	"flag"
	"log"
	"os"

	"github.com/fatih/color"
	"google.golang.org/genai"
)

var (
	defaultBackend = genai.BackendGeminiAPI
	defaultModel   = "gemini-2.0-flash"

	bluef     = color.New(color.FgHiBlue).PrintfFunc()
	greenf    = color.New(color.FgHiGreen).PrintfFunc()
	greenln   = color.New(color.FgHiGreen).PrintlnFunc()
	magentaf  = color.New(color.FgHiMagenta).PrintfFunc()
	magentaln = color.New(color.FgHiMagenta).PrintlnFunc()
	redf      = color.New(color.FgHiRed).PrintfFunc()
	redln     = color.New(color.FgHiRed).PrintlnFunc()
	reds      = color.New(color.FgHiRed).SprintFunc()
	whitef    = color.New(color.FgHiWhite).PrintfFunc()
	whiteln   = color.New(color.FgHiWhite).PrintlnFunc()
	yellowf   = color.New(color.FgHiYellow).PrintfFunc()
	yellowln  = color.New(color.FgHiYellow).PrintlnFunc()

	// Command line flags.
	logSession = flag.Bool("log_session", true, "If true, the conversation with the tool is logged to a file in the location set in the CODE_AGENT_LOGDIR environment (or in /tmp when CODE_AGENT_LOGDIR is not set")
)

// GetUserMessage captures the user input in the terminal.
func GetUserMessage(s *bufio.Scanner) (string, bool) {
	if !s.Scan() {
		return "", false
	}
	return s.Text(), true
}

func main() {
	flag.Parse()

	ctx := context.Background()
	// The APIKey field will be read from the environment GOOGLE_API_KEY.
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend: defaultBackend,
	})
	if err != nil {
		log.Fatalf("Unable to create a genai client: %s", reds(err.Error()))
	}
	greenln("client created \u2714")

	tools := []*genai.Tool{readFileTool, editFileTool, listFilesTool}
	systemInstr := &genai.Content{
		Parts: []*genai.Part{
			{Text: "Answer concisely. Ask clarifying questions, if necessary."},
		},
		Role: "user",
	}
	config := &genai.GenerateContentConfig{
		SystemInstruction: systemInstr,
		CandidateCount:    1,
		Tools:             tools,
	}
	model := defaultModel
	chat, err := client.Chats.Create(ctx, model, config, nil)
	if err != nil {
		log.Fatalf("Unable to create a chat session: %s\n", reds(err.Error()))
	}
	greenf("chat with '%s' created \u2714\n", model)

	getter := func() (string, bool) {
		return GetUserMessage(bufio.NewScanner(os.Stdin))
	}

	var logFile *os.File
	if *logSession {
		logFile = LogFileOrNil(LogFilename("session"))
	} else {
		whiteln("session not logged")
	}
	if logFile != nil {
		whitef("logging conversation to '%s'\n", logFile.Name())
	}

	agent := NewAgent(client, model, getter, logFile)
	err = agent.Run(context.Background(), chat)
	if err != nil {
		log.Fatalf("unable to run agent: %s\n", reds(err.Error()))
	}
}
