#!/bin/bash

set -e

protoc --proto_path=. --go_out=. ./pkg/protocol/buy_offers.proto
protoc --proto_path=. --go_out=. ./pkg/protocol/invoices.proto
protoc --proto_path=. --go_out=. ./pkg/protocol/mint.proto
protoc --proto_path=. --go_out=. ./pkg/protocol/payment.proto
protoc --proto_path=. --go_out=. ./pkg/protocol/sell_offers.proto

