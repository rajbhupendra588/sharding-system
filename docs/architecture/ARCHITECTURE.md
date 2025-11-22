# System Architecture

## Overview
The Sharding System is designed to provide scalable and reliable data storage.

## Components
-   **Router**: Routes requests to the appropriate shard.
-   **Manager**: Manages shard metadata and lifecycle.
-   **Shard**: Stores the actual data.

## Data Flow
1.  Client sends request to Router.
2.  Router queries Manager for shard location.
3.  Router forwards request to Shard.
4.  Shard processes request and returns response.
