package cats

// RespObj defines API /fact response.
type RespObj struct {
	Fact   string `json:"fact"`
	Length int    `json:"length"`
}
