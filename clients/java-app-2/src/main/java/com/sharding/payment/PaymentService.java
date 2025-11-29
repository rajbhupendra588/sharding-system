package com.sharding.payment;

import com.fasterxml.jackson.databind.ObjectMapper;
import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.sql.*;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import com.sun.net.httpserver.HttpServer;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpExchange;

public class PaymentService {
    private static Connection db;
    private static ObjectMapper mapper = new ObjectMapper();

    public static void main(String[] args) throws Exception {
        String dbHost = System.getenv().getOrDefault("DB_HOST", "localhost");
        String dbPort = System.getenv().getOrDefault("DB_PORT", "5432");
        String dbUser = System.getenv().getOrDefault("DB_USER", "postgres");
        String dbPassword = System.getenv().getOrDefault("DB_PASSWORD", "postgres");
        String dbName = System.getenv().getOrDefault("DB_NAME", "payments_db");

        String url = String.format("jdbc:postgresql://%s:%s/%s", dbHost, dbPort, dbName);
        db = DriverManager.getConnection(url, dbUser, dbPassword);

        initDB();

        HttpServer server = HttpServer.create(new InetSocketAddress(8080), 0);
        server.createContext("/health", new HealthHandler());
        server.createContext("/api/payments", new PaymentsHandler());
        server.createContext("/api/payments/stats", new StatsHandler());
        server.setExecutor(null);
        server.start();
        System.out.println("Payment service started on port 8080");
    }

    private static void initDB() throws SQLException {
        String createTable = """
            CREATE TABLE IF NOT EXISTS payments (
                id SERIAL PRIMARY KEY,
                order_id INTEGER NOT NULL,
                amount DECIMAL(10,2) NOT NULL,
                payment_method VARCHAR(50) NOT NULL,
                status VARCHAR(50) NOT NULL,
                transaction_id VARCHAR(100),
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            );
            
            CREATE INDEX IF NOT EXISTS idx_payments_order_id ON payments(order_id);
            CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
            CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments(transaction_id);
            """;

        try (Statement stmt = db.createStatement()) {
            stmt.execute(createTable);
        }

        // Insert sample data
        try (PreparedStatement checkStmt = db.prepareStatement("SELECT COUNT(*) FROM payments");
             ResultSet rs = checkStmt.executeQuery()) {
            rs.next();
            if (rs.getInt(1) == 0) {
                String insert = "INSERT INTO payments (order_id, amount, payment_method, status, transaction_id) VALUES (?, ?, ?, ?, ?)";
                try (PreparedStatement pstmt = db.prepareStatement(insert)) {
                    Object[][] samples = {
                        {1, 199.98, "credit_card", "completed", "txn_001"},
                        {2, 29.99, "paypal", "completed", "txn_002"},
                        {3, 239.97, "credit_card", "pending", "txn_003"},
                        {4, 299.99, "debit_card", "completed", "txn_004"},
                        {5, 299.98, "credit_card", "completed", "txn_005"}
                    };
                    for (Object[] sample : samples) {
                        pstmt.setInt(1, (Integer) sample[0]);
                        pstmt.setDouble(2, (Double) sample[1]);
                        pstmt.setString(3, (String) sample[2]);
                        pstmt.setString(4, (String) sample[3]);
                        pstmt.setString(5, (String) sample[4]);
                        pstmt.executeUpdate();
                    }
                }
            }
        }
    }

    static class HealthHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange exchange) throws IOException {
            try {
                db.createStatement().executeQuery("SELECT 1");
                Map<String, String> response = Map.of("status", "healthy");
                sendJsonResponse(exchange, 200, response);
            } catch (SQLException e) {
                sendJsonResponse(exchange, 503, Map.of("status", "unhealthy", "error", e.getMessage()));
            }
        }
    }

    static class PaymentsHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange exchange) throws IOException {
            if ("GET".equals(exchange.getRequestMethod())) {
                handleGet(exchange);
            } else if ("POST".equals(exchange.getRequestMethod())) {
                handlePost(exchange);
            } else {
                exchange.sendResponseHeaders(405, -1);
            }
        }

        private void handleGet(HttpExchange exchange) throws IOException {
            try {
                List<Map<String, Object>> payments = new ArrayList<>();
                try (PreparedStatement stmt = db.prepareStatement("SELECT id, order_id, amount, payment_method, status, transaction_id, created_at FROM payments ORDER BY id");
                     ResultSet rs = stmt.executeQuery()) {
                    while (rs.next()) {
                        Map<String, Object> payment = new HashMap<>();
                        payment.put("id", rs.getInt("id"));
                        payment.put("order_id", rs.getInt("order_id"));
                        payment.put("amount", rs.getDouble("amount"));
                        payment.put("payment_method", rs.getString("payment_method"));
                        payment.put("status", rs.getString("status"));
                        payment.put("transaction_id", rs.getString("transaction_id"));
                        payment.put("created_at", rs.getTimestamp("created_at").toString());
                        payments.add(payment);
                    }
                }
                sendJsonResponse(exchange, 200, payments);
            } catch (SQLException e) {
                sendJsonResponse(exchange, 500, Map.of("error", e.getMessage()));
            }
        }

        private void handlePost(HttpExchange exchange) throws IOException {
            try {
                String body = new String(exchange.getRequestBody().readAllBytes());
                Map<String, Object> payment = mapper.readValue(body, Map.class);
                
                String sql = "INSERT INTO payments (order_id, amount, payment_method, status, transaction_id) VALUES (?, ?, ?, ?, ?) RETURNING id, created_at";
                try (PreparedStatement stmt = db.prepareStatement(sql)) {
                    stmt.setInt(1, (Integer) payment.get("order_id"));
                    stmt.setDouble(2, ((Number) payment.get("amount")).doubleValue());
                    stmt.setString(3, (String) payment.get("payment_method"));
                    stmt.setString(4, (String) payment.get("status"));
                    stmt.setString(5, (String) payment.get("transaction_id"));
                    
                    try (ResultSet rs = stmt.executeQuery()) {
                        rs.next();
                        payment.put("id", rs.getInt("id"));
                        payment.put("created_at", rs.getTimestamp("created_at").toString());
                    }
                }
                sendJsonResponse(exchange, 201, payment);
            } catch (Exception e) {
                sendJsonResponse(exchange, 400, Map.of("error", e.getMessage()));
            }
        }
    }

    static class StatsHandler implements HttpHandler {
        @Override
        public void handle(HttpExchange exchange) throws IOException {
            try {
                Map<String, Object> stats = new HashMap<>();
                try (PreparedStatement stmt = db.prepareStatement("""
                    SELECT 
                        COUNT(*) as total_payments,
                        COALESCE(SUM(amount), 0) as total_amount,
                        COALESCE(AVG(amount), 0) as avg_amount,
                        COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed_count
                    FROM payments
                    """); ResultSet rs = stmt.executeQuery()) {
                    if (rs.next()) {
                        stats.put("total_payments", rs.getInt("total_payments"));
                        stats.put("total_amount", rs.getDouble("total_amount"));
                        stats.put("avg_amount", rs.getDouble("avg_amount"));
                        stats.put("completed_count", rs.getInt("completed_count"));
                    }
                }
                sendJsonResponse(exchange, 200, stats);
            } catch (SQLException e) {
                sendJsonResponse(exchange, 500, Map.of("error", e.getMessage()));
            }
        }
    }

    private static void sendJsonResponse(HttpExchange exchange, int statusCode, Object response) throws IOException {
        String json = mapper.writeValueAsString(response);
        exchange.getResponseHeaders().set("Content-Type", "application/json");
        exchange.sendResponseHeaders(statusCode, json.getBytes().length);
        try (OutputStream os = exchange.getResponseBody()) {
            os.write(json.getBytes());
        }
    }
}

