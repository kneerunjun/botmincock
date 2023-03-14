package main

type BotUpdate struct {
	Id      int64 `json:"update_id"`
	Message struct {
		From struct {
			Id    int64  `json:"id"`
			UName string `json:"username"`
		} `json:"from"`
		Text string `json:"text"`
	} `json:"message"`
}
