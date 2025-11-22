# System Design

## Design Principles
-   **Scalability**: Horizontal scaling of shards.
-   **Availability**: Replication and failover.
-   **Consistency**: Strong consistency for metadata, eventual consistency for data (configurable).

## Technologies
-   **Language**: Go
-   **Storage**: PostgreSQL (Metadata), Custom (Shards)
-   **Communication**: gRPC, REST
