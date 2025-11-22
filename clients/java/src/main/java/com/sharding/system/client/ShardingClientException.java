package com.sharding.system.client;

/**
 * Exception thrown by ShardingClient when operations fail.
 */
public class ShardingClientException extends Exception {
    public ShardingClientException(String message) {
        super(message);
    }
    
    public ShardingClientException(String message, Throwable cause) {
        super(message, cause);
    }
}


