package com.sharding.order;

import com.fasterxml.jackson.databind.ObjectMapper;
import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.sql.*;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import com.sun.net.httpserver.HttpServer;
import com.sun.net.httpserver.HttpHandler;
import com.sun.net.httpserver.HttpExchange;

public class OrderService {
    private static Connection db;
    private static ObjectMapper mapper = new ObjectMapper();

    public static void main(String[] args) throws Exception {
        String dbHost = System.getenv().getOrDefault("DB_HOST", "localhost");
        String dbPort = System.getenv().getOrDefault("DB_PORT", "5432");
        String dbUser = System.getenv().getOrDefault("DB_USER", "postgres");
        String dbPassword = System.getenv().getOrDefault("DB_PASSWORD", "postgres");
        String dbName = System.getenv().getOrDefault("DB_NAME", "orders_db");

        String url = String.format("jdbc:postgresql://%s:%s/%s", dbHost, dbPort, dbName);
        db = DriverManager.getConnection(url, dbUser, dbPassword);

        initDB();

        HttpServer server = HttpServer.create(new InetSocketAddress(8080), 0);
        server.createContext("/health", new HealthHandler());
        server.createContext("/api/orders", new OrdersHandler());
        server.createContext("/api/orders/stats", new StatsHandler());
        server.setExecutor(null);
        server.start();
        System.out.println("Order service started on port 8080");
    }

    private static void initDB() throws SQLException {
        String createTable = """
            CREATE TABLE IF NOT EXISTS orders (
                id SERIAL PRIMARY KEY,
                customer_id INTEGER NOT NULL,
                product_id INTEGER NOT NULL,
                quantity INTEGER NOT NULL,
                total_amount DECIMAL(10,2) NOT NULL,
                status VARCHAR(50) NOT NULL,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            );
            
            CREATE INDEX IF NOT EXISTS idx_orders_customer_id ON orders(customer_id);
            CREATE INDEX IF NOT EXISTS idx_orders_status ON orders(status);
            """;

        try (Statement stmt = db.createStatement()) {
            stmt.execute(createTable);
        }

        // Insert sample data
        try (PreparedStatement checkStmt = db.prepareStatement("SELECT COUNT(*) FROM orders");
             ResultSet rs = checkStmt.executeQuery()) {
            rs.next();
            if (rs.getInt(1) == 0) {
                String insert = "INSERT INTO orders (customer_id, product_id, quantity, total_amount, status) VALUES (?, ?, ?, ?, ?)";
                try (PreparedStatement pstmt = db.prepareStatement(insert)) {
                    Object[][] samples = {
                        {1, 101, 2, 199.98, "pending"},
                        {2, 102, 1, 29.99, "completed"},
                        {3, 103, 3, 239.97, "pending"},
                        {4, 104, 1, 299.99, "shipped"},
                        {5, 105, 2, 299.98, "completed"}
                    };
                    for (Object[] sample : samples) {
                        pstmt.setInt(1, (Integer) sample[0]);
                        pstmt.setInt(2, (Integer) sample[1]);
                        pstmt.setInt(3, (Integer) sample[2]);
                        pstmt.setDouble(4, (Double) sample[3]);
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

    static class OrdersHandler implements HttpHandler {
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
                List<Map<String, Object>> orders = new ArrayList<>();
                try (PreparedStatement stmt = db.prepareStatement("SELECT id, customer_id, product_id, quantity, total_amount, status, created_at FROM orders ORDER BY id");
                     ResultSet rs = stmt.executeQuery()) {
                    while (rs.next()) {
                        Map<String, Object> order = new HashMap<>();
                        order.put("id", rs.getInt("id"));
                        order.put("customer_id", rs.getInt("customer_id"));
                        order.put("product_id", rs.getInt("product_id"));
                        order.put("quantity", rs.getInt("quantity"));
                        order.put("total_amount", rs.getDouble("total_amount"));
                        order.put("status", rs.getString("status"));
                        order.put("created_at", rs.getTimestamp("created_at").toString());
                        orders.add(order);
                    }
                }
                sendJsonResponse(exchange, 200, orders);
            } catch (SQLException e) {
                sendJsonResponse(exchange, 500, Map.of("error", e.getMessage()));
            }
        }

        private void handlePost(HttpExchange exchange) throws IOException {
            try {
                String body = new String(exchange.getRequestBody().readAllBytes());
                Map<String, Object> order = mapper.readValue(body, Map.class);
                
                String sql = "INSERT INTO orders (customer_id, product_id, quantity, total_amount, status) VALUES (?, ?, ?, ?, ?) RETURNING id, created_at";
                try (PreparedStatement stmt = db.prepareStatement(sql)) {
                    stmt.setInt(1, (Integer) order.get("customer_id"));
                    stmt.setInt(2, (Integer) order.get("product_id"));
                    stmt.setInt(3, (Integer) order.get("quantity"));
                    stmt.setDouble(4, ((Number) order.get("total_amount")).doubleValue());
                    stmt.setString(5, (String) order.get("status"));
                    
                    try (ResultSet rs = stmt.executeQuery()) {
                        rs.next();
                        order.put("id", rs.getInt("id"));
                        order.put("created_at", rs.getTimestamp("created_at").toString());
                    }
                }
                sendJsonResponse(exchange, 201, order);
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
                        COUNT(*) as total_orders,
                        COALESCE(SUM(total_amount), 0) as total_revenue,
                        COALESCE(AVG(total_amount), 0) as avg_order_value
                    FROM orders
                    """); ResultSet rs = stmt.executeQuery()) {
                    if (rs.next()) {
                        stats.put("total_orders", rs.getInt("total_orders"));
                        stats.put("total_revenue", rs.getDouble("total_revenue"));
                        stats.put("avg_order_value", rs.getDouble("avg_order_value"));
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

