#!/bin/bash

# This file is for local development only!
# It configures and starts the website for local development.

export MB_LOGLEVEL=debug
export MB_ALLOWED_ORIGINS=*
export MB_HOST=localhost:8080

go run main.go
