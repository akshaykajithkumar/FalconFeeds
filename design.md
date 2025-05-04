# FalconFeeds Design

## Assumptions
- The system is designed to collect publicly available OSINT feeds and process them into STIX 2.1 format.
- Redis serves as a buffer to decouple the feed collection and normalization services, allowing for resilient, asynchronous processing.
- MongoDB is used for persistent storage of normalized indicators.
- The system uses Docker and Docker Compose to simplify deployment and ensure consistency between environments.

## System Components

### 1. **Feed Collector**
- **Goal**: Fetch raw OSINT data at regular intervals and push it into Redis.
- **Design Considerations**:
  - Polling interval of 5 minutes is sufficient for this use case, balancing performance and resource utilization.
  - OpenTelemetry is integrated for tracing feed fetches and identifying bottlenecks.
  - Error handling is crucial, ensuring that feed failures are logged and retried.
  
### 2. **Normalizer**
- **Goal**: Process raw feed data, extract IOCs, and normalize them into STIX 2.1 format.
- **Design Considerations**:
  - The system uses regular expressions to extract IOCs (IP, domain, SHA-256).
  - The normalizer uses STIX 2.1 standard objects: Indicator, ObservedData, Relationship, and Observable.
  - MongoDB is used for persistence, enabling efficient querying of normalized indicators.
  - The API allows for flexible IOC searches by value or ID, but further query optimization can be made.

## Assumptions and Trade-offs

- **Feed Reliability**: We are assuming that the external feeds are reliable. If a feed is down, the system will log the failure and retry in subsequent pollings.
- **Performance**: The system is designed to handle moderate feed traffic but may need optimization if scaling is necessary (e.g., multi-threaded feed collection, caching).
- **Data Latency**: The feed polling interval of 5 minutes is a trade-off between data freshness and resource utilization.

### What I’d Improve with More Time
- Explore using Kafka for more scalable and durable stream processing.
- Develop a web-based UI for visualizing STIX indicators
-  Prometheus + Grafana
- Add more complex querying capabilities (e.g., searching by multiple indicators at once).
- Integrate additional external threat intelligence sources.

---

## Conclusion

The FalconFeeds project is designed to provide a modular, scalable approach to OSINT feed collection and normalization. By using Docker Compose for deployment and Redis for decoupling services, the architecture ensures flexibility and resilience. Further improvements could focus on scalability, IOC enrichment, and advanced analytics.
