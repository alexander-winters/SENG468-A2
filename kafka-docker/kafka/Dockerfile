# Use the official Kafka image as the base image
FROM confluentinc/cp-kafka:latest

# Set environment variables
ENV KAFKA_BROKER_ID 1
ENV KAFKA_ZOOKEEPER_CONNECT zookeeper:2181
ENV KAFKA_ADVERTISED_LISTENERS PLAINTEXT://kafka:9092
ENV KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR 1

# Expose the default Kafka broker port
EXPOSE 9092