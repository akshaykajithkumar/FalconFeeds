# FalconFeeds - OSINT Indicator Normalization System

This is a modular system designed to collect public feeds, normalize them into STIX 2.1 indicator objects, and make them available for analysis or further consumption. It contains two services: **Feed Collector** and **Normalizer**.

---

## Services Overview

### 1. **Feed Collector** 

- **Goal**: Collects pulls raw threat-intel feeds from publicly available sources, processes them every 5 minutes, and publishes raw data to a Redis stream.
- **Feed Sources Used For Testing**:
  - Malware Bazaar RSS
  - CERT Bund Advisories
  - PhishTank JSON feed (for phishing-related data)
- **Output**: The raw feed data is sent to the `raw-feeds` Redis stream.

### 2. **Normalizer**

- **Goal**: Consumes raw feed data from the Redis stream, extracts Indicators of Compromise (IOCs) such as IP addresses, domains, and file hashes, and normalizes them into STIX 2.1 format.
- **STIX Objects**:
  - **Indicator**
  - **ObservedData**
  - **Relationship**
  - **Observable (IP, File, Domain)**
- **Output**: The normalized STIX data is published to the `stix-indicators` Redis stream and stored in MongoDB.

---

## Features

- **Polling**: Feeds are polled every 5 minutes.
- **STIX 2.1 Compliance**: Transforms raw data into STIX 2.1 objects for standardized threat intelligence exchange.
- **Resilience**: Graceful shutdown and context-aware Redis/MongoDB interactions.
- **Docker-Compose**: One command to start the entire system (`make up`).
- **Health Check API**: Each service exposes a `/healthz` endpoint.
- **IOC Query API**: The Normalizer exposes an API to query IOCs by value (e.g., IP, domain, SHA-256).

---

## Architecture Diagram
          +--------------------+        
        |   Feed Collector   |        
        |   (Service A)      |        
        +--------------------+        
                |                    
                v                    
      +------------------------+      
      |    Redis Streams       |      
      |   (raw-feeds stream)   |      
      +------------------------+      
                |                    
                v                    
        +-------------------+        
        |  Normalizer       |        
        |  (Service B)     |        
        +-------------------+        
                |                    
                v                    
         +--------------+            
         |   MongoDB   |            
         | (stix-indicators) |
         +--------------+            
                |                    
                v                    
          +-------------+             
          | Prometheus |             
          | (Metrics)  |             
          +-------------+             
                |                    
                v                    
          +-------------+             
          |   Jaeger   |             
          | (Tracing)  |             
          +-------------+             

---

---

## How to Run

1. **Clone the repository**:
    ```bash
    git clone https://github.com/akshaykajithkumar/FalconFeeds.git
    cd falconfeeds
    ```

2. **Set up the environment** (make sure Docker is installed):
    ```bash
    make up
    ```

3. **services**:
   - **Feed Collector**:( Runs automatically, fetching and pushing feeds every 5 minutes.
   - **Normalizer**: Starts processing raw feeds from Redis, normalizing them into STIX 2.1 objects.
   
4. **Access Prometheus**:
   - Prometheus metrics are available at `http://localhost:9090`.

5. **Access Jaeger**:
   - Jaeger traces are available on the Jaeger UI: `http://localhost:16686`.


---

## **API Endpoints**

### **1. Health Check**

- **GET /healthz (Feed Collector)**:
   - URL: `http://localhost:4000/healthz`
   - Returns the health status of the Feed Collector service.

- **GET /healthz (Normalizer)**:
   - URL: `http://localhost:5000/healthz`
   - Returns the health status of the Normalizer service.

### **2. IOC Query API**

- **GET /indicators**:
   - URL: `http://localhost:5000/indicators?value=<value>&limit=<limit>`
   - **Description**: Queries IOCs in normalized STIX 2.1 format.
   - **Query Parameters**:
     - `value`: The IOC value to search for (e.g., an IP address, domain, or SHA-256 hash).
     - `limit`: The number of results to return (default: 10).

   - **Example**: Search for indicators with a SHA-256 hash value:
     ```
     http://localhost:5000/indicators?value=hash&limit=10
     ```

## **Testing**
- **Unit tests** have been added to verify the functionality of individual components and ensure they behave as expected.
- End-to-end **integration tests** have been implemented to validate the entire flow of the system, ensuring that all components work seamlessly together.

  ### Setting Up the Test Environment
  `docker run -d -p 6379:6379 --name test-redis redis`

  `docker run -d -p 27017:27017 --name test-mongo mongo`
---

---

## Dependencies

- **Redis**: Used for message queuing between services.
- **MongoDB**: Used for storing normalized STIX 2.1 data.
- **Prometheus**: Used for monitoring and collecting metrics from services.
- **Jaeger**: Used for distributed tracing, enabling detailed performance analysis and request flow tracking.


---
