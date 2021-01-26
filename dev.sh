#!/bin/bash

# This file is for local development only!
# It configures and starts the website for local development.
# The "secret" for the Emvi API can be shared, as it gives access to public content only.

export MB_LOGLEVEL=debug
export MB_ALLOWED_ORIGINS=*
export MB_HOST=localhost:8080
export MB_HOT_RELOAD=true
export MB_EMVI_CLIENT_ID=3fBBn144yvSF9R3dPC8l
export MB_EMVI_CLIENT_SECRET=
export MB_EMVI_ORGA=marvin
export MB_PIRSCH_CLIENT_ID=gEb3pvgxZvZzFRlOTdMgPtyLvNYgeVKe
export MB_PIRSCH_CLIENT_SECRET=E7UqJehmxgnVuw81oq6ZhJAx9vCHqMimCUFfil7UFgbGhgQVVINqU7JqHBgaUvHg
export MB_PIRSCH_HOSTNAME=marvinblum.de

go run main.go
