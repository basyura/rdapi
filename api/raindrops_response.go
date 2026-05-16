package api

type RaindropsResponse struct {
	Result bool       `json:"result"`
	Items  []Raindrop `json:"items"`
	Count  int        `json:"count"`
}
