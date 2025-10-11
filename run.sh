#!/bin/bash
if [ "$1" = 'down' ]; then
  docker compose -f ./docker/compose.dev.yaml down
elif [ "$1" = 'build' ]; then
  docker compose -f ./docker/compose.dev.yaml build
else
  docker compose -f ./docker/compose.dev.yaml up --build
fi
