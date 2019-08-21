#!/bin/bash
set -e

echo "┌──────────────────────────────────────────┐"
echo "│                                          │"
echo "│        Running Website Tests...          │"
echo "│                                          │"
echo "└──────────────────────────────────────────┘"

# Build the container
make website

# Run the tests
docker-compose -f dpc-web/docker-compose.yml run web rails db:create db:migrate db:seed
docker-compose -f dpc-web/docker-compose.yml run web rails spec
docker-compose -f dpc-web/docker-compose.yml down

echo "┌──────────────────────────────────────────┐"
echo "│                                          │"
echo "│        All Website Tests Complete        │"
echo "│                                          │"
echo "└──────────────────────────────────────────┘"