package models

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Result any `json:"result"`
}
