#!/bin/bash

echo "Clearing docker logs..."
echo "" > $(docker inspect --format='{{.LogPath}}' postgres)
