# Use the official Zookeeper image as the base image
FROM zookeeper:latest

# Set environment variables
ENV ZOO_TICK_TIME 2000
ENV ZOO_INIT_LIMIT 5
ENV ZOO_SYNC_LIMIT 2

# Expose the default Zookeeper client port
EXPOSE 2181