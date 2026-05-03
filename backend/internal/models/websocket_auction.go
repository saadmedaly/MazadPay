package models

 

// WSEvent — tous les événements WebSocket envoyés aux clients
type WSEvent struct {
	Type    string      `json:"type"`    // bid_placed | timer_tick | auction_ended | auction_won
	Payload interface{} `json:"payload"`
}

type BidPlacedPayload struct {
	AuctionID    string  `json:"auction_id"`
	NewPrice     float64 `json:"new_price"`
	BidderMasked string  `json:"bidder_phone"` // "####4709"
	BidCount     int     `json:"bid_count"`
	SecondsLeft  int64   `json:"seconds_left"`
}

type TimerTickPayload struct {
	AuctionID   string `json:"auction_id"`
	SecondsLeft int64  `json:"seconds_left"`
}

type AuctionEndedPayload struct {
	AuctionID   string  `json:"auction_id"`
	FinalPrice  float64 `json:"final_price"`
	WinnerPhone string  `json:"winner_phone"` // masqué
}
