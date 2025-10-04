package mcp_server //nolint:revive,stylecheck // fine for now

type GreetingInput struct {
	Name string `json:"name" jsonschema:"the name of the person to greet"`
}

type GreetingOutput struct {
	Greeting string `json:"greeting" jsonschema:"the greeting to tell to the user"`
}
