#!/usr/bin/env bash

go run ./cmd/harnejrd --listen "${HARNEJR_LISTEN:-127.0.0.1:8765}" --config-dir configs
