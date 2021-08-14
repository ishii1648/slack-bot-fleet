package example

type RequestBody struct {
	Reaction      string `json:"reaction"`
	User          string `json:"user"`
	ItemChannel   string `json:"itemChannel"`
	ItemTimestamp string `json:"itemTimestamp"`
}
