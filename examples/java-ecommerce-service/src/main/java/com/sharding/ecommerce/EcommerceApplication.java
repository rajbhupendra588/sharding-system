package com.sharding.ecommerce;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.boot.context.properties.EnableConfigurationProperties;

/**
 * Main Spring Boot Application for E-Commerce Service with Database Sharding.
 * 
 * This application demonstrates how to use the Sharding System to:
 * - Scale horizontally by distributing data across multiple database shards
 * - Handle high-volume transactions efficiently
 * - Maintain data consistency and availability
 * - Perform seamless resharding operations
 */
@SpringBootApplication
@EnableConfigurationProperties
public class EcommerceApplication {

    public static void main(String[] args) {
        SpringApplication.run(EcommerceApplication.class, args);
    }
}

