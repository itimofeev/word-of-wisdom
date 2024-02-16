package entity

type RequestChallengeMessage struct {
}

type ResponseChallengeMessage struct {
	Data       []byte `json:"data"`
	Difficulty uint   `json:"difficulty"`
}

type SolvedChallengeMessage struct {
	Nonce []byte `json:"nonce"`
}

type QuoteMessage struct {
	Quote string `json:"quote"`
}
