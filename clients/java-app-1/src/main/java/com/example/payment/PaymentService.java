package com.example.payment;

import com.google.gson.Gson;
import spark.Request;
import spark.Response;
import spark.Spark;

import java.sql.*;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class PaymentService {
    private static Connection db;
    private static final Gson gson = new Gson();

    public static void main(String[] args) {
        initDatabase();
        setupRoutes();
    }

    private static void initDatabase() {
        String dbHost = System.getenv().getOrDefault("DB_HOST", "localhost");
        String dbPort = System.getenv().getOrDefault("DB_PORT", "5432");
        String dbUser = System.getenv().getOrDefault("DB_USER", "postgres");
        String dbPassword = System.getenv().getOrDefault("DB_PASSWORD", "postgres");
        String dbName = System.getenv().getOrDefault("DB_NAME", "payment_db");

        String url = String.format("jdbc:postgresql://%s:%s/%s", dbHost, dbPort, dbName);

        try {
            db = DriverManager.getConnection(url, dbUser, dbPassword);
            System.out.println("Connected to database successfully");

            String createTable = """
                CREATE TABLE IF NOT EXISTS payments (
                    id SERIAL PRIMARY KEY,
                    order_id INTEGER NOT NULL,
                    amount DECIMAL(10,2) NOT NULL,
                    payment_method VARCHAR(50) NOT NULL,
                    status VARCHAR(50) DEFAULT 'pending',
                    transaction_id VARCHAR(255),
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );
                """;

            try (Statement stmt = db.createStatement()) {
                stmt.execute(createTable);
            }

            String insertData = """
                INSERT INTO payments (order_id, amount, payment_method, status, transaction_id)
                VALUES 
                    (1, 1999.98, 'credit_card', 'completed', 'txn_001'),
                    (2, 149.95, 'paypal', 'pending', 'txn_002'),
                    (3, 79.99, 'debit_card', 'completed', 'txn_003')
                ON CONFLICT DO NOTHING;
                """;

            try (Statement stmt = db.createStatement()) {
                stmt.execute(insertData);
            }

            System.out.println("Database initialized successfully");
        } catch (SQLException e) {
            System.err.println("Database connection failed: " + e.getMessage());
            System.exit(1);
        }
    }

    private static void setupRoutes() {
        Spark.port(Integer.parseInt(System.getenv().getOrDefault("PORT", "8080")));

        Spark.get("/health", PaymentService::healthCheck);
        Spark.get("/payments", PaymentService::getPayments);

        System.out.println("Payment service started on port " + System.getenv().getOrDefault("PORT", "8080"));
    }

    private static String healthCheck(Request req, Response res) {
        try {
            if (db == null || db.isClosed()) {
                res.status(503);
                return "Database connection failed";
            }
            db.createStatement().executeQuery("SELECT 1");
            res.status(200);
            return "OK";
        } catch (SQLException e) {
            res.status(503);
            return "Database connection failed: " + e.getMessage();
        }
    }

    private static String getPayments(Request req, Response res) {
        List<Map<String, Object>> payments = new ArrayList<>();
        try {
            String query = "SELECT id, order_id, amount, payment_method, status, transaction_id, created_at FROM payments";
            try (Statement stmt = db.createStatement();
                 ResultSet rs = stmt.executeQuery(query)) {
                while (rs.next()) {
                    Map<String, Object> payment = new HashMap<>();
                    payment.put("id", rs.getInt("id"));
                    payment.put("order_id", rs.getInt("order_id"));
                    payment.put("amount", rs.getBigDecimal("amount"));
                    payment.put("payment_method", rs.getString("payment_method"));
                    payment.put("status", rs.getString("status"));
                    payment.put("transaction_id", rs.getString("transaction_id"));
                    payment.put("created_at", rs.getTimestamp("created_at").toString());
                    payments.add(payment);
                }
            }
            res.type("application/json");
            return gson.toJson(Map.of("payments", payments));
        } catch (SQLException e) {
            res.status(500);
            return "{\"error\": \"" + e.getMessage() + "\"}";
        }
    }
}

