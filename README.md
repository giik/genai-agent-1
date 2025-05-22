# An Interactive, Code-editing Agent using Gemini

## Inspired by [How to Build and Agent](https://ampcode.com/how-to-build-an-agent)

Wow, what an [amazing write
up](https://ampcode.com/how-to-build-an-agent) by [Thorsten
Ball](https://www.linkedin.com/in/thorsten-ball-3142b652)! I got
really inspired to try it out but I could not: turns out, one needs
credits for Claude that I wasn't willing to pay for. But I do have
Google Gemini so I asked myself: "how difficult could it be to borrow
the code and twist it to work with the Google API?" Well, it took
longer than I expected but I'm still giving myself credit for doing it
in a less than a day after not having written any Go for about 10
years, and, having zero familiarity with the Google [Gemini API
Libraries](https://ai.google.dev/gemini-api/docs/libraries), incl. the
Go version.


What are we making here: a command-line tool that connects
to Gemini and allows it to edit code on our (the client) side.

## Initial Setup Required

You gotta have Go on your system and that's easy to achieve. Follow
the instructions on [Go installation](https://go.dev/doc/install) for
your platform. If you have an older version-as I did-just uninstall it
by following the [Uninstalling
Go](https://go.dev/doc/manage-install#uninstalling) guide, then
reinstall the newest version.

Next, head to Google's [AI
Studio](https://aistudio.google.com/app/apikey) and get an API
key. Click on the Create API Key button in the top right corner, then
click in the Search Google Cloud projects field of the popup (I didn't
have any existing Cloud Projects) and select Gemini API. Copy and save
the key (you won't be able to view it again) then make sure it's in
your command line environment. I use z shell on macos; you can follow
[these
instructions](https://ai.google.dev/gemini-api/docs/api-key#macos---zsh)
if you're completely new to this. Once the key is exported it will be
picked up from the environment (so you don't have to&mdash;and you do
**not** want to&mdash;hardcode it into the program). I used
`GOOGLE_API_KEY` for the environment variable name and it works.

Clone the [repo at
github.com/giik/genai-agent-1](https://github.com/giik/genai-agent-1). To
build and run the code use the following command:

```bash
go build ./... && ./agent
```

## Notes

Again, this code uses the Google GenAI API from
`google.golang.org/genai` so we just need to import it. Similar to
Anthropic, the GenAI API provides for multi-turn conversations with
tool/function calling.

The idea is very elegant and there is an excellent, [easily
understandable description (picture
included)](https://ai.google.dev/gemini-api/docs/function-calling?example=meeting#how_function_calling_works)
in the Gemini API docs. In short, the system prompt (sent to the model
from our client at the start of the session) includes the definitions
of the available functions/tools.  The model can decide to answer with
a Text response, or, with a "function call" to one of the tools we've
made available. Then, it is our responsibility to run the tool-client
side-and forward any output back to the model, which send back its
final answer. Very ingenious!

To tie it all together, we need a few pieces:

* write the tools we need, such as a `ReadFile` function,
* let Gemini know of the available tools using the `Tools` field of
the `GenerateContentConfig`; this includes the tool name and
description, and (most importantly) its schema, such as inputs, and
output,
* respond to Gemini's requests to run tools by calling the
corresponding functions with the Gemini-provided parameters, then
forwarding outputs to Gemini.

Also, note that on every turn we send to the model a single text part
or one or more responses to the Function Call requests. The server is
stateless, so the actual requests to it must carry the context the
model needs -- typically, all turns in the conversation so far -- but
this is handled by the API, not our client code.  I am guessing there
is some point where the context becomes too large and gets truncated
but this is just a wild guess, I don't really know and haven't yet
inspected the GenAI Go code to see whether it actually does that.

## Logging 

I added functionality to save the conversation to a markdown file
because the model's responses are in markdown and that makes it easy
to incorporate them. Log files are named like
`session-20250518-1150-a1bd959eaa.md` so they'll sort nicely if you
need to find one ot the most recent one quickly. The log files are
written to the location specified in the `CODE_AGENT_LOGDIR`
environment variable which I personally set to
`$HOME/Developer/agentlogs` (or to `/tmp` if the environment is not
found). Logging can be turned off completely using the `log_session`
command line flag like

```bash
go build ./... && ./agent -log_session=false
```

## Firing it up

When you fire it up, things should look like this:

```bash
go build ./... && ./agent
client created ✔
chat with 'gemini-2.0-flash' created ✔
logging conversation to '/Users/****/Developer/agentlog/session-20250521-2301-d1e8345801.md'
[user]: Hello, I am Bob.
[model][text 24 bytes][part 1 of 1] Hi Bob, how can I help?

[user]: What tools are available to you?
[model][text 82 bytes][part 1 of 1] I have access to the following tools: `read_file`, `edit_file`, and `list_files`.

[user]:
```

I recorded a session (see the included file
`session-20250520-2142-d2e9e2fbfb.md`) in which I asked the model to
create a fizzbuzz Swift program, then edit it to change how far it's
going (the program is included).  This worked great. Similarly, it was
able to create a functional SwiftUI view (included) of a
timer/clock. Finally, because I couldn't help myself I did ask it to
read its code (all Go files in the current directory) and write its
own markdown documentation. I've included it without changes, it's not
at all bad.

## About Go

The last time I wrote a relatively complex server project in Go (or
any code in Go) was about 2017. It was a nice break from the daily C++
at work and whatever other hobbies at home. It came back to me in a
jiffy but I ask the more experienced Gophers out there to forgive me
for any non-idiomatic use or otherwise unorthodox style.

Enjoy!

Model Generated Documentation Following
---------------------------------------

# Code Agent Architecture: Detailed Analysis

This document provides an in-depth analysis of the code agent's architecture and functionality, based on the provided Go source code.

## Core Components

*   **`main.go`**: The main entry point. It initializes the Google Gemini API client, configures the chat session, and starts the agent's main loop.

    *   It reads the `GOOGLE_API_KEY` environment variable for authentication.
    *   It defines command-line flags, specifically `log_session` to enable/disable logging.
    *   It initializes the Gemini API client using `genai.NewClient`.
    *   It registers available tools (`readFileTool`, `editFileTool`, `listFilesTool`) with the Gemini model.
    *   It sets up a system instruction to guide the LLM.  Critically, this instruction tells the model how to behave ("Answer concisely. Ask clarifying questions, if necessary.").
    *   It creates a new chat session using `client.Chats.Create`.
    *   It creates an `Agent` instance with the Gemini client, model name, a function to get user input (`GetUserMessage`), and a log file (if logging is enabled).
    *   Finally, it calls `agent.Run` to start the agent's main loop.

*   **`agent.go`**: Contains the `Agent` struct and the main control loop (`Run` method).

    *   The `Agent` struct holds the Gemini client, model name, a function to obtain user input (`GetPrompt`), and a log file.
    *   The `Run` method implements a state machine to manage the conversation flow.  The states are `UserInput`, `SendMessage`, `ProcessResponse`, and `Error`.
    *   **UserInput**: Prompts the user for input using the `GetPrompt` function.
    *   **SendMessage**: Sends the user's input to the Gemini API using `chat.SendMessage`.
    *   **ProcessResponse**: Handles the response from the Gemini API.
        *   It iterates through the response parts, handling both text and function calls.
        *   For function calls, it extracts the function name and arguments and calls the appropriate function using the `FunctionCall` function (defined in `function_call.go`).
        *   It appends the results of the function call (either the output or an error message) to the conversation.
        *   It logs the conversation to the log file using the `AppendToLog` function.
    *   The `DescribeMapEntries` function converts a map[string]any into a slice of `EntryInfo` structs, making it easier to describe the contents of the map in the logs. It uses reflection to provide detailed descriptions of different data types.

*   **`function_call.go`**: Handles the execution of function calls requested by the LLM.

    *   The `FunctionCall` function takes a `genai.FunctionCall` struct as input.
    *   It validates the function name and arguments against a list of expected tools and arguments.
    *   It uses a `switch` statement to call the appropriate function based on the function name (`read_file`, `edit_file`, or `list_files`).
    *   It constructs a `genai.FunctionResponse` struct containing the result of the function call (either the output or an error message).
    *   It uses `greenf` and `redf` to indicate success or failure of a function call.

*   **`edit_file.go`**: Implements the `edit_file` tool.

    *   The `EditFile` function takes the file path, the old string, and the new string as input.
    *   It reads the content of the file using the `ReadFile` function.
    *   It replaces all occurrences of the old string with the new string using `strings.Replace`.
    *   It writes the modified content back to the file using `os.WriteFile`.
    *   It includes error handling for cases where the file does not exist or the old string is not found.
    *   The `CreateNewFile` function creates a new file with the specified content, including creating any necessary parent directories.

*   **`read_file.go`**: Implements the `read_file` tool.

    *   The `ReadFile` function takes the file path as input.
    *   It reads the content of the file using `os.ReadFile` and returns it as a string.

*   **`list_files.go`**: Implements the `list_files` tool.

    *   The `ListFiles` function takes the path as input.
    *   It uses `filepath.WalkDir` to recursively traverse the directory specified by the path.
    *   It creates a list of all files and directories within the specified path.
    *   It marshals the list of files into a JSON string using `json.Marshal`.

*   **`logging.go`**: Provides logging functionality.

    *   The `LogFileOrNil` function creates a log file in the directory specified by the `CODE_AGENT_LOGDIR` environment variable (or `/tmp` if the environment variable is not set). The filename includes a timestamp and a random string.
    *   The `AppendToLog` function appends a message to the log file, including the role of the message sender (user or model) and the timestamp.

## Data Flow

1.  The user provides input to the agent via the command line.
2.  The `main.go` file reads the user's input and sends it to the Gemini API.
3.  The Gemini API processes the input and generates a response.
4.  If the response contains a function call, the `agent.go` file calls the appropriate function using the `function_call.go` file.
5.  The function call interacts with the file system (using `edit_file.go`, `read_file.go`, or `list_files.go`).
6.  The result of the function call is sent back to the Gemini API.
7.  The Gemini API generates a final response.
8.  The `main.go` file displays the response to the user.
9.  The entire conversation is logged to a file using `logging.go`.

## Security Considerations

The agent implements several security measures:

*   **Tool Validation**: The `function_call.go` validates that the requested tool exists and that the arguments are correct.
*   **Sandboxing (Implicit)**: The agent relies on the Gemini API to provide a sandboxed environment for code execution. While the provided code does not explicitly implement sandboxing, the interaction with external resources (file system) are mediated through function calls defined and validated by the agent.
*   **Limited File System Access**: The agent only provides access to a limited set of file system operations (read, edit, list).  Direct shell execution or arbitrary code execution is not supported.

## Improvements

*   **Explicit Sandboxing**: Implementing an explicit sandboxing mechanism (e.g., using containers or virtual machines) would further enhance security.
*   **Access Control**: Implementing more fine-grained access control policies would allow for more secure management of file system resources.
*   **Rate Limiting**: Implementing rate limiting would prevent the agent from being overwhelmed with requests.
*   **Input Sanitization**: Implementing more robust input sanitization would prevent injection attacks.

This detailed analysis provides a comprehensive understanding of the code agent's architecture, functionality, and security considerations.  It highlights the key components, data flow, and potential areas for improvement.


