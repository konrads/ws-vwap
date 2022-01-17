package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/konrads/ws-vwap/pkg/algo"
	"github.com/konrads/ws-vwap/pkg/buffer"
	"github.com/konrads/ws-vwap/pkg/model"
	log "github.com/sirupsen/logrus"
)

// WS client based on: https://golangdocs.com/golang-gorilla-websockets
func main() {
	wsUrl := flag.String(`coinbaseUrl`, `ws-feed-public.sandbox.exchange.coinbase.com`, `coinbase ws address`)
	productIDsStr := flag.String(`product-ids`, `BTC-USD,ETH-USD,ETH-BTC`, `comma separated products`)
	keepDubious := flag.Bool(`keep-dubious`, false, `keep datapoints where eg. price is set to 100.00`)
	vwapWindowSize := flag.Uint(`vwap-window-size`, 200, `size of the VWAP sliding window`)
	logLevel := flag.Uint(`log-level`, uint(log.InfoLevel), fmt.Sprintf("logrus levels, eg debug: %d, info: %d, warn: %d", log.DebugLevel, log.InfoLevel, log.WarnLevel))
	flag.Parse()

	log.SetLevel(log.Level(*logLevel))
	productIDs := strings.Split(*productIDsStr, ",")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: `wss`, Host: *wsUrl}
	log.Infof("connecting to %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatalf("failed to establish WS connection due to: %v", err)
	}

	done := make(chan struct{})
	go mainLoop(conn, productIDs, *keepDubious, *vwapWindowSize)

	defer close(done)
	defer conn.Close()

	// await done/interrupt prior to exiting
	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Info(`received interrupt, closing`)
			err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Fatal("closing error:", err)
			}
			// await done message or timeout
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

// Main worker, parses the WS messsages, pushes them onto RingBuffer and calculates VWAPs
func mainLoop(conn *websocket.Conn, productIDs []string, keepDubious bool, vwapWindowSize uint) {
	byProductBuffer := map[string]buffer.RingBuffer{}
	subscribeJson, _ := json.Marshal(model.SubscribeMsg{Type: `subscribe`, ProductIDs: productIDs, Channels: []string{`level2`}})
	log.Info(`sending subscribe`, string(subscribeJson))
	conn.WriteMessage(websocket.TextMessage, subscribeJson)
	for {
		_, msgBytes, err := conn.ReadMessage()
		if err != nil {
			log.Fatal(`error reading ws message:`, err)
		}
		msg, err := model.ParseMsg(msgBytes)
		if err != nil {
			log.Fatalf("error parsing message %v, %s", err, string(msgBytes))
		}
		switch m := msg.(type) {
		case model.TypedUpdateMsg:
			log.Debug("update:", msg)
			for _, change := range m.Changes {
				// only process non dubious entries
				if keepDubious || !isDubious(change) {
					buff, ok := byProductBuffer[m.ProductID]
					if !ok {
						b, err := buffer.NewRingBuffer(vwapWindowSize)
						if err != nil {
							log.Fatal(`failed to create RingBuffer:`, err)
						}
						byProductBuffer[m.ProductID] = b
						buff = b
					}
					buff.Push(buffer.PriceVol{Price: change.Price, Vol: change.Qty})
					vwap, err := algo.VWAP(buff)
					if err != nil {
						log.Fatal(`error calculating vwap:`, err)
					}
					log.Infof("VWAP for %s: %f", m.ProductID, vwap)
				}
			}
		case model.SubscriptionsMsg:
			log.Info("subscription:", msg)
		case model.ErrorMsg:
			log.Warn("error:", msg)
		case model.SnapshotMsg:
			log.Debug("snapshot:", msg)
		default:
			log.Debugf("unexpected msg: %s", m)
		}
	}
}

func isDubious(c model.UpdateChange) bool {
	return c.Price == 100.0
}
