package ws

import (
    "time"

    "github.com/gofiber/websocket/v2"
    "go.uber.org/zap"
)

const (
    writeWait      = 10 * time.Second
    pongWait       = 60 * time.Second
    pingPeriod     = (pongWait * 9) / 10
    maxMessageSize = 512
)

type Client struct {
    conn   *websocket.Conn
    send   chan []byte
    userID string
    Role   string // Admin role (admin, super_admin) for admin connections
    market string // Customer #20 global channel: connecting user's EffectiveAccountCountryISO(), used by GlobalHub.BroadcastAuctionEvent's market filter. Empty for per-auction/admin clients, which don't need it.
    logger *zap.Logger
}

func NewClient(conn *websocket.Conn, userID string, logger *zap.Logger) *Client {
    return &Client{
        conn:   conn,
        send:   make(chan []byte, 64),
        userID: userID,
        logger: logger,
    }
}

// SetMarket stamps the connecting user's market onto the client (Customer
// #20 global channel only) -- exported so HandleGlobal can set it without
// needing a market-aware constructor variant.
func (c *Client) SetMarket(market string) {
    c.market = market
}

// WritePump envoie les messages en attente au client WebSocket
func (c *Client) WritePump() {
    ticker := time.NewTicker(pingPeriod)
    defer func() {
        ticker.Stop()
        c.conn.Close()
    }()

    for {
        select {
        case message, ok := <-c.send:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if !ok {
                c.conn.WriteMessage(websocket.CloseMessage, []byte{})
                return
            }
            if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
                return
            }

        case <-ticker.C:
            c.conn.SetWriteDeadline(time.Now().Add(writeWait))
            if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
                return
            }
        }
    }
}

// ReadPump lit les messages entrants (ping/pong keepalive)
func (c *Client) ReadPump() {
    defer c.conn.Close()
    c.conn.SetReadLimit(maxMessageSize)
    c.conn.SetReadDeadline(time.Now().Add(pongWait))
    c.conn.SetPongHandler(func(string) error {
        c.conn.SetReadDeadline(time.Now().Add(pongWait))
        return nil
    })
    for {
        if _, _, err := c.conn.ReadMessage(); err != nil {
            break
        }
    }
}
