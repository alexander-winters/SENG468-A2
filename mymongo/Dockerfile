FROM mongo:latest

# Set environment variables for MongoDB
ENV MONGO_INITDB_DATABASE=seng468-a2-db

# Copy the current directory contents into the container at /usr/src/app
COPY db.go /docker-entrypoint-initdb.d/

# Expose port 27017 for MongoDB
EXPOSE 27017

# Start MongoDB when the container launches
CMD ["mongod"]