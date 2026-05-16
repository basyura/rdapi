package api

type Raindrop struct {
	ID      int    `json:"_id"`
	Title   string `json:"title"`
	Link    string `json:"link"`
	Domain  string `json:"domain"`
	Created string `json:"created"`
}
