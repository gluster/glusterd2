package api

// Webhook is Structure to represent a webhook that will be used
// for posting events
type Webhook struct {
	URL    string `json:"url"`
	Token  string `json:"token"`
	Secret string `json:"secret"`
}

// WebhookDel is Structure to represent a webhook that will be used
// for deleting webhook
type WebhookDel struct {
	URL string `json:"url"`
}
