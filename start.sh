#!/bin/bash

STACK_NAME="opsmon"
COMPOSE_FILE="pipeline/docker-compose.yml"

BACKEND_DIR="server"
FRONTEND_DIR="frontend"
ALERTS_FILE="server/backend/engine/alerts.json"

start_stack() {

    echo "Starting Elasticsearch..."
    docker compose -f $COMPOSE_FILE up -d

    echo "Waiting for Elasticsearch..."
    sleep 10

    echo "Starting Backend..."
    cd $BACKEND_DIR
    go build -o opsmon cmd/main.go
    ./opsmon &
    BACK_PID=$!
    echo $BACK_PID > .backend.pid
    cd ..

    echo "Starting Frontend..."
    cd $FRONTEND_DIR
    npm install --silent
    npm run dev &
    FRONT_PID=$!
    echo $FRONT_PID > .frontend.pid
    cd ..

    echo "OpsMon Demo Running"
}

stop_stack() {

    echo "Stopping Backend..."
    kill $(cat server/.backend.pid) 2>/dev/null

    echo "Stopping Frontend..."
    kill $(cat frontend/.frontend.pid) 2>/dev/null

    echo "Stopping Elasticsearch..."
    docker compose -f $COMPOSE_FILE down
}

soft_reset() {

    echo "Clearing Alerts..."
    # echo "[]" > $ALERTS_FILE

    echo "Clearing Elasticsearch Index..."
    curl -X DELETE localhost:9200/logger

    echo "Soft Reset Done"
}

hard_reset() {

    echo "Hard Reset Initiated..."

    stop_stack

    echo "Removing Node Modules..."
    rm -rf frontend/node_modules

    echo "Removing Go Build..."
    rm -rf server/opsmon

    echo "Removing Alerts..."
    echo "[]" > $ALERTS_FILE

    echo "Removing ES Containers..."
    docker compose -f $COMPOSE_FILE down -v

    echo "Removing Volumes..."
    docker volume prune -f

    echo "Removing Networks..."
    docker network prune -f

    echo "Pulling Fresh Images..."
    docker compose -f $COMPOSE_FILE pull

    start_stack
}

logs_stack() {
    docker compose -f $COMPOSE_FILE logs -f
}

case "$1" in
    start) start_stack ;;
    stop) stop_stack ;;
    softreset) soft_reset ;;
    hardreset) hard_reset ;;
    logs) logs_stack ;;
    *)
        echo "Usage: $0 {start|stop|softreset|hardreset|logs}"
        ;;
esac
