#!/bin/bash
set -e

TOPIC=${KAFKA_TOPIC:-robogo_test}
BROKER=${KAFKA_BROKER:-kafka:9092}

echo "Waiting for Kafka broker at $BROKER to be available..."

for i in {1..20}; do
  if kafka-topics.sh --bootstrap-server "$BROKER" --list > /dev/null 2>&1; then
    echo "Kafka broker is available, creating topic $TOPIC..."
    kafka-topics.sh --create --if-not-exists --topic "$TOPIC" --bootstrap-server "$BROKER" --partitions 1 --replication-factor 1
    echo "Kafka topic '$TOPIC' created or already exists."
    exit 0
  else
    echo "Kafka not ready yet, retrying in 5s... ($i/20)"
    sleep 5
  fi
done

echo "Failed to connect to Kafka broker after multiple attempts."
exit 1