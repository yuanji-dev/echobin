package main

type responseRef struct {
	URL     string            `json:"url"`
	Args    interface{}       `json:"args"`
	Form    interface{}       `json:"form"`
	Data    interface{}       `json:"data"`
	Origin  string            `json:"origin"`
	Headers map[string]string `json:"headers"`
	Files   interface{}       `json:"files"`
	JSON    interface{}       `json:"json"`
	Method  string            `json:"method"`
}

type getResponse struct {
	Args    map[string]interface{} `json:"args"`
	Headers map[string]string      `json:"headers"`
	Origin  string                 `json:"origin"`
	URL     string                 `json:"url"`
}
