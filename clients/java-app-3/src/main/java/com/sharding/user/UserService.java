package com.sharding.user;

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

public class UserService {
    private static Connection db;
    private static ObjectMapper mapper = new ObjectMapper();

    public static void main(String[] args) throws Exception {
        String dbHost = System.getenv().getOrDefault("DB_HOST", "localhost");
        String dbPort = System.getenv().getOrDefault("DB_PORT", "5432");
        String dbUser = System.getenv().getOrDefault("DB_USER", "postgres");
        String dbPassword = System.getenv().getOrDefault("DB_PASSWORD", "postgres");
        String dbName = System.getenv().getOrDefault("DB_NAME", "users_db");

        String url = String.format("jdbc:postgresql://%s:%s/%s", dbHost, dbPort, dbName);
        db = DriverManager.getConnection(url, dbUser, dbPassword);

        initDB();

        HttpServer server = HttpServer.create(new InetSocketAddress(8080), 0);
        server.createContext("/health", new HealthHandler());
        server.createContext("/api/users", new UsersHandler());
        server.createContext("/api/users/stats", new StatsHandler());
        server.setExecutor(null);
        server.start();
        System.out.println("User service started on port 8080");
    }

    private static void initDB() throws SQLException {
        String createTable = """
            CREATE TABLE IF NOT EXISTS users (
                id SERIAL PRIMARY KEY,
                username VARCHAR(100) NOT NULL UNIQUE,
                email VARCHAR(255) NOT NULL UNIQUE,
                first_name VARCHAR(100),
                last_name VARCHAR(100),
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
                last_login TIMESTAMP
            );
            
            CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
            CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
            """;

        try (Statement stmt = db.createStatement()) {
            stmt.execute(createTable);
        }

        // Insert sample data
        try (PreparedStatement checkStmt = db.prepareStatement("SELECT COUNT(*) FROM users");
             ResultSet rs = checkStmt.executeQuery()) {
            rs.next();
            if (rs.getInt(1) == 0) {
                String insert = "INSERT INTO users (username, email, first_name, last_name) VALUES (?, ?, ?, ?)";
                try (PreparedStatement pstmt = db.prepareStatement(insert)) {
                    Object[][] samples = {
                        {"john_doe", "john@example.com", "John", "Doe"},
                        {"jane_smith", "jane@example.com", "Jane", "Smith"},
                        {"bob_wilson", "bob@example.com", "Bob", "Wilson"},
                        {"alice_brown", "alice@example.com", "Alice", "Brown"},
                        {"charlie_davis", "charlie@example.com", "Charlie", "Davis"}
                    };
                    for (Object[] sample : samples) {
                        pstmt.setString(1, (String) sample[0]);
                        pstmt.setString(2, (String) sample[1]);
                        pstmt.setString(3, (String) sample[2]);
                        pstmt.setString(4, (String) sample[3]);
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

    static class UsersHandler implements HttpHandler {
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
                List<Map<String, Object>> users = new ArrayList<>();
                try (PreparedStatement stmt = db.prepareStatement("SELECT id, username, email, first_name, last_name, created_at, last_login FROM users ORDER BY id");
                     ResultSet rs = stmt.executeQuery()) {
                    while (rs.next()) {
                        Map<String, Object> user = new HashMap<>();
                        user.put("id", rs.getInt("id"));
                        user.put("username", rs.getString("username"));
                        user.put("email", rs.getString("email"));
                        user.put("first_name", rs.getString("first_name"));
                        user.put("last_name", rs.getString("last_name"));
                        Timestamp createdAt = rs.getTimestamp("created_at");
                        if (createdAt != null) user.put("created_at", createdAt.toString());
                        Timestamp lastLogin = rs.getTimestamp("last_login");
                        if (lastLogin != null) user.put("last_login", lastLogin.toString());
                        users.add(user);
                    }
                }
                sendJsonResponse(exchange, 200, users);
            } catch (SQLException e) {
                sendJsonResponse(exchange, 500, Map.of("error", e.getMessage()));
            }
        }

        private void handlePost(HttpExchange exchange) throws IOException {
            try {
                String body = new String(exchange.getRequestBody().readAllBytes());
                Map<String, Object> user = mapper.readValue(body, Map.class);
                
                String sql = "INSERT INTO users (username, email, first_name, last_name) VALUES (?, ?, ?, ?) RETURNING id, created_at";
                try (PreparedStatement stmt = db.prepareStatement(sql)) {
                    stmt.setString(1, (String) user.get("username"));
                    stmt.setString(2, (String) user.get("email"));
                    stmt.setString(3, (String) user.get("first_name"));
                    stmt.setString(4, (String) user.get("last_name"));
                    
                    try (ResultSet rs = stmt.executeQuery()) {
                        rs.next();
                        user.put("id", rs.getInt("id"));
                        user.put("created_at", rs.getTimestamp("created_at").toString());
                    }
                }
                sendJsonResponse(exchange, 201, user);
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
                        COUNT(*) as total_users,
                        COUNT(CASE WHEN last_login IS NOT NULL THEN 1 END) as active_users,
                        COUNT(CASE WHEN created_at > CURRENT_DATE - INTERVAL '30 days' THEN 1 END) as new_users_30d
                    FROM users
                    """); ResultSet rs = stmt.executeQuery()) {
                    if (rs.next()) {
                        stats.put("total_users", rs.getInt("total_users"));
                        stats.put("active_users", rs.getInt("active_users"));
                        stats.put("new_users_30d", rs.getInt("new_users_30d"));
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

