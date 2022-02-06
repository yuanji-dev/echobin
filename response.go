package main

type responseRef struct {
	Args    map[string]interface{} `json:"args"`
	Data    string                 `json:"data"`
	Files   map[string]interface{} `json:"files"`
	Form    interface{}            `json:"form"`
	Headers map[string]string      `json:"headers"`
	JSON    map[string]interface{} `json:"json"`
	Method  string                 `json:"method"`
	Origin  string                 `json:"origin"`
	URL     string                 `json:"url"`
}

type getMethodResponse struct {
	Args    map[string]interface{} `json:"args"`
	Headers map[string]string      `json:"headers"`
	Origin  string                 `json:"origin"`
	URL     string                 `json:"url"`
}

type otherMethodResponse struct {
	Args    map[string]interface{} `json:"args"`
	Data    string                 `json:"data"`
	Files   map[string]interface{} `json:"files"`
	Form    interface{}            `json:"form"`
	Headers map[string]string      `json:"headers"`
	JSON    map[string]interface{} `json:"json"`
	Origin  string                 `json:"origin"`
	URL     string                 `json:"url"`
}

type requestHeadersResponse struct {
	Headers map[string]string `json:"headers"`
}

type requestIPResponse struct {
	Origin string `json:"origin"`
}

type requestUserAgentResponse struct {
	UserAgent string `json:"user-agent"`
}

type encodedResponse struct {
	Origin  string            `json:"origin"`
	Headers map[string]string `json:"headers"`
	Method  string            `json:"method"`
}

type gzippedResponse struct {
	encodedResponse
	Gzipped bool `json:"gzipped"`
}

type deflatedResponse struct {
	encodedResponse
	Deflated bool `json:"deflated"`
}

type delayResponse struct {
	Args    map[string]interface{} `json:"args"`
	Data    string                 `json:"data"`
	Files   map[string]interface{} `json:"files"`
	Form    interface{}            `json:"form"`
	Headers map[string]string      `json:"headers"`
	Origin  string                 `json:"origin"`
	URL     string                 `json:"url"`
}

type streamResponse struct {
	Args    map[string]interface{} `json:"args"`
	Headers map[string]string      `json:"headers"`
	Origin  string                 `json:"origin"`
	URL     string                 `json:"url"`
	ID      int                    `json:"id"`
}
