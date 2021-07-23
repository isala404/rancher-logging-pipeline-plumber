package webserver

type HTTPError struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
	Info    string `json:"info,omitempty"`
}

type HTTPData struct {
	Message string      `json:"message"`
	Info    interface{} `json:"info,omitempty"`
}

type HTTPResponse struct {
	Error *HTTPError `json:"error,omitempty"`
	Data  *HTTPData  `json:"data,omitempty"`
}
