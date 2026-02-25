#!/bin/bash

STACK_NAME="opsmon"
COMPOSE_FILE="docker-compose.yml"

start_stack() {
    echo "🚀 Starting OpsMon Stack..."
    docker compose -f $COMPOSE_FILE up -d
}

stop_stack() {
    echo "🛑 Stopping OpsMon Stack..."
    docker compose -f $COMPOSE_FILE down
}

restart_stack() {
    echo "🔄 Restarting OpsMon Stack..."
    docker compose -f $COMPOSE_FILE down
    docker compose -f $COMPOSE_FILE up -d
}

reset_stack() {
    echo "🔥 HARD RESET (Removing ONLY storage volumes)..."
    
    docker compose -f $COMPOSE_FILE down -v
    
    echo "🧹 Removing dangling volumes..."
    docker volume prune -f

    echo "🚀 Starting Fresh Containers (Images NOT Re-pulled)..."
    docker compose -f $COMPOSE_FILE up -d
}

status_stack() {
    docker ps | grep opsmon
}

logs_stack() {
    docker compose -f $COMPOSE_FILE logs -f
}


full_reset_stack() {
    echo "💀 FULL HARD RESET (Containers + Volumes + Images + Network)..."

    echo "🛑 Bringing down stack..."
    docker compose -f $COMPOSE_FILE down -v --remove-orphans

    echo "🧹 Removing OpsMon containers..."
    docker ps -a --filter "name=$STACK_NAME" -q | xargs -r docker rm -f

    echo "🧹 Removing OpsMon volumes..."
    docker volume ls -q --filter "name=$STACK_NAME" | xargs -r docker volume rm

    echo "🧹 Removing OpsMon network..."
    docker network ls -q --filter "name=$STACK_NAME" | xargs -r docker network rm

    echo "🧹 Removing Elasticsearch & Kibana images..."
    docker images docker.elastic.co/elasticsearch/elasticsearch --format "{{.ID}}" | xargs -r docker rmi -f

    echo "🧹 Pruning build cache..."
    docker builder prune -f

    echo "🚀 Pulling fresh images..."
    docker compose -f $COMPOSE_FILE pull

    echo "🚀 Starting fresh stack..."
    docker compose -f $COMPOSE_FILE up -d
}

case "$1" in
    start) start_stack ;;
    stop) stop_stack ;;
    restart) restart_stack ;;
    reset) reset_stack ;;
    status) status_stack ;;
    logs) logs_stack ;;
    fullreset) full_reset_stack ;;
    *)
        echo "Usage: $0 {start|stop|restart|reset|status|logs}"
        ;;
esac
