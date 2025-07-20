# Example Tests

This directory contains simple, illustrative test cases for documentation, onboarding, and quick demos.

## How to Run

```bash
./robogo.exe run examples/
``` 

## Create Kafka Topic

docker exec kafka kafka-topics.sh --create --topic test-topic --bootstrap-server localhost:9092 --partitions 1 --replication-factor 1