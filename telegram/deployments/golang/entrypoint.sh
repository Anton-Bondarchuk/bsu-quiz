#!/bin/sh

if [ "$ENV" = "local" ]; then
  exec /app/main --config="./config/local.yml"
else
  exec /app/main
fi