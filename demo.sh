#!/usr/bin/env bash

set -euo pipefail

./build/big-iot-gateway-darwin start \
  --config config.yaml \
  --debug \
  --offerFile offers.json \
  --offeringCheckIntervalSec 30 \
  --offeringEndpoint http://538e0ed7.ngrok.io \
  "$@"