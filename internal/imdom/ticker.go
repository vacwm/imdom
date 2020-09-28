package imdom

import (
	"errors"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
	"github.com/vacwm/go-rapi"
	"github.com/vacwm/go-rapi/pkg/market"
	"google.golang.org/protobuf/proto"
)

// TickerPlant is stateful connection to Rithmic's TickerPlant service
type TickerPlant struct {
	conn *websocket.Conn

	// IsClosed returns true if the connection closed
	IsClosed bool

	// OrderbookSub specifies symbols subscribed for orderbook updates.
	OrderBookSubscription map[string]chan market.OrderBook

	// TradeSub specifies symbols subscribed to trade events.
	TradeSubscription map[string]chan market.LastTrade

	// DepthSub specifies symbols subscribed to depth by order updates.
	DepthSubscription map[string]chan market.DepthByOrder
}

// NewTickerPlant creates a TickerPlant instance. The method may fail while establishing
// a connection.
func NewTickerPlant(url url.URL) (*TickerPlant, error) {
	log.Printf("Connecting to %s", url.String())
	conn, _, err := websocket.DefaultDialer.Dial(url.String(), nil)
	if err != nil {
		return nil, err
	}

	// Create login request
	conn.WriteMessage(websocket.BinaryMessage, requestLogin())

	// Read login response (blocking)
	_, binary, err := conn.ReadMessage()
	if err != nil {
		return nil, err
	}
	templateID, err := DecodeBytes(binary)
	if err != nil {
		return nil, err
	}
	switch templateID {
	case 11:
		responseLogin := rti.ResponseLogin{}
		proto.Unmarshal(binary[4:], &responseLogin)
		if responseLogin.RpCode[0] != "0" {
			return nil, errors.New(fmt.Sprintf("Login request failed with error: %s\n", responseLogin.RpCode[1]))
		}
	}

	return &TickerPlant{
		conn:                  conn,
		IsClosed:              false,
		OrderBookSubscription: make(map[string]chan market.OrderBook),
		TradeSubscription:     make(map[string]chan market.LastTrade),
		DepthSubscription:     make(map[string]chan market.DepthByOrder),
	}, nil
}

// Run starts a loop to process incoming events
func (t *TickerPlant) Run() {
	for {
		_, binary, err := t.conn.ReadMessage()
		if err != nil {
			log.Println("There was an error processing TickerPlant message: ", err)
			return
		}
		templateID, err := DecodeBytes(binary)
		if err != nil {
			log.Println("TickerPlant error while decoding binary: ", err)
			return
		}
		switch templateID {
		case 19:
			log.Println("Heartbeat Response received")
		case 101:
			response := market.ResponseMarketDataUpdate{}
			proto.Unmarshal(binary[4:], &response)
			log.Printf("Received market data update response: %s\n", &response)
		case 156:
			response := market.OrderBook{}
			proto.Unmarshal(binary[4:], &response)
			t.OrderBookSubscription[*response.Symbol] <- response
		default:
			log.Printf("Unrecognized templateID: %d\n", templateID)
		}
	}
}

func requestLogin() []byte {
	request := &rti.RequestLogin{}
	templateID := int32(10)
	appVersion := "1.0"
	request.TemplateId = &templateID
	request.AppVersion = &appVersion
	message, err := proto.Marshal(request)
	if err != nil {
		log.Panicln(err)
	}
	return EncodeByteLength(message)
}

// Close closes the connection.
func (t *TickerPlant) Close() error {
	t.conn.Close()
	t.IsClosed = true
	return nil
}

// SubscribeTrade returns stream of Trade events for the provided symbol/exchange pair.
func (t *TickerPlant) SubscribeTrade(symbol string, exchange string) chan market.LastTrade {
	if stream, ok := t.TradeSubscription[symbol]; !ok {
		// Create channel
		t.TradeSubscription[symbol] = make(chan market.LastTrade)

		// Create subscription request
		templateID := int32(100)
		updateBits := uint32(market.RequestMarketDataUpdate_LAST_TRADE)
		request := market.RequestMarketDataUpdate_SUBSCRIBE

		requestMarketDataUpdate := &market.RequestMarketDataUpdate{
			TemplateId: &templateID,
			Symbol:     &symbol,
			Exchange:   &exchange,
			UpdateBits: &updateBits,
			Request:    &request,
		}
		message, err := proto.Marshal(requestMarketDataUpdate)
		if err != nil {
			log.Panicln(err)
		}

		// Send request
		t.conn.WriteMessage(websocket.BinaryMessage, EncodeByteLength(message))
		return t.TradeSubscription[symbol]
	} else {
		return stream
	}
}

// SubscribeOrderBook returns stream of OrderBook updates for the provided symbol/exchange pair.
func (t *TickerPlant) SubscribeOrderBook(symbol string, exchange string) chan market.OrderBook {
	if stream, ok := t.OrderBookSubscription[symbol]; !ok {
		// Create channel
		t.OrderBookSubscription[symbol] = make(chan market.OrderBook)

		// Create subscription request
		templateID := int32(100)
		updateBits := uint32(market.RequestMarketDataUpdate_ORDER_BOOK)
		request := market.RequestMarketDataUpdate_SUBSCRIBE

		requestMarketDataUpdate := &market.RequestMarketDataUpdate{
			TemplateId: &templateID,
			Symbol:     &symbol,
			Exchange:   &exchange,
			UpdateBits: &updateBits,
			Request:    &request,
		}
		message, err := proto.Marshal(requestMarketDataUpdate)
		if err != nil {
			log.Panicln(err)
		}

		// Send request
		t.conn.WriteMessage(websocket.BinaryMessage, EncodeByteLength(message))
		return t.OrderBookSubscription[symbol]
	} else {
		return stream
	}
}
