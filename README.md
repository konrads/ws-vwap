# Websocket coinbase VWAP calculator

* connects to Coinbase WS API via `level2` channel. Note, this channel provides volume hence was preferred to `ticker`
* creates a circular `RingBuffer` per each product
* per new `l2update` event, pushes new price/volume into the buffer, reads back `VWAP`

## To run:
```bash
# test
go test -v ./...
# run WS client
go run cmd/ws_client/main.go
# or in debug mode
go run cmd/ws_client/main.go -log-level 5
```

## Background research
Events were observed with the help of `wscat` utility
```bash
# WS connection
wscat --connect wss://ws-feed-public.sandbox.exchange.coinbase.com
# channel subscription
> { "type": "subscribe", "product_ids": [ "BTC-USD", "ETH-USD", "ETH-BTC" ], "channels": [ "level2" ] }
# incoming WS messages
< {"type":"error","message":"Failed to subscribe","reason":"ETH-USD is not a valid product"}
< {"type":"subscriptions","channels":[{"name":"level2","product_ids":["ETH-BTC","BTC-USD"]}]}
< {"type":"snapshot","product_id":"ETH-BTC","asks":[["0.07600","311.83164253"],["0.07900","0.29910000"]........
< {"type":"l2update","product_id":"BTC-USD","changes":[["sell","45632.17","522.05252813"]],"time":"2022-01-17T02:57:26.530613Z"}
< {"type":"l2update","product_id":"BTC-USD","changes":[["buy","100.00","26.04800000"]],"time":"2022-01-17T02:57:31.034898Z"}
```

## Snags
* as per output above, seems `level2` subscription doesn't accept `ETH-USD` product - program reports the error and processes other products.
* occasionally coinbase responds with following price data, where price is set to `100.00` and volume is large. This skews the VWAP and is ignored by default. To unable, use `-keep-dubious` flag:
```{"type":"l2update","product_id":"BTC-USD","changes":[["buy","100.00","26.04800000"]],"time":"2022-01-17T02:57:31.034898Z"}```