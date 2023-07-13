# Redis
Redis is an open-source, in-memory data structure store that can be used as a database, cache, and message broker. It provides high performance, scalability, and flexibility, making it a popular choice for various use cases.

# init
The InitRedis function in the database package initializes and connects to a Redis database based on the provided configuration. It establishes a connection to the Redis server and returns a Redis client that can be used for interacting with the Redis database.

## Redis Initialization
The InitRedis function in the database package is responsible for initializing and connecting to a Redis database based on the provided Redis configuration.

## Redis Client Initialization
The InitRedis function takes a redisConfig parameter of type *config.Redis, which contains the Redis configuration parameters such as host, port, user, password and TTL. It creates a new Redis client using the redis.NewClient function with the provided configuration options.

## Connecting to Redis
After creating the Redis client, the Ping method is called on the client to establish a connection to the Redis server. If the connection is successful, the Ping method returns a "PONG" response. If there is an error connecting to Redis, the function panics with an error message.

# Mock redis 
The NewRedisMock function creates a mock Redis server and a Redis client for testing purposes. This allows for testing Redis-related functionality without the need for a real Redis server

## Redis Mock Initialization
The NewRedisMock function in the database package is responsible for creating a mock Redis server and a Redis client for testing purposes.

## Mock Redis Server
The NewRedisMock function uses the miniredis.Run function to start a mock Redis server. The mock server is a lightweight in-memory Redis server that can be used for testing without the need for a real Redis server.

## Redis Client for Mock Server
After starting the mock Redis server, the function creates a Redis client using the redis.NewClient function and the address of the mock server obtained from server.Addr().

# Redis usage in this project 
The Flight handler in the project package is responsible for handling flight-related requests. It utilizes Redis for caching flight data and improving performance. The Following documentation will guide you through the Redis usage in the Flight handler.

## Caching Flight Data
The Get method in the Flight handler retrieves flight data from the Redis cache if available. It uses a cache key generated based on the request parameters, such as departure city, arrival city, and flight date. The cache key is used to retrieve the cached flight data from Redis using the Get method of the Redis client.
If the flight data is not found in the cache (indicated by the redis.Nil error), the flight data is fetched from an external API using the APIMock client. The fetched flight data is then stored (for a time specified by TTL in config) in the Redis cache using the Set method of the Redis client, with the cache key and a JSON representation of the flight data as the parameters.

## Redis Error Handling
In case of any errors during Redis operations, appropriate error responses are returned to the client. For example, if there is an error retrieving flight data from the Redis cache or storing flight data in the cache, the Get method returns an internal server error response.
