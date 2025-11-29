package com.sharding;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.sql.*;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class InventoryService {
    private static Connection db;
    private static ObjectMapper mapper = new ObjectMapper();

    public static void main(String[] args) throws Exception {
        // Database connection
        String dbHost = System.getenv().getOrDefault("DB_HOST", "localhost");
        String dbPort = System.getenv().getOrDefault("DB_PORT", "5432");
        String dbUser = System.getenv().getOrDefault("DB_USER", "postgres");
        String dbPassword = System.getenv().getOrDefault("DB_PASSWORD", "postgres");
        String dbName = System.getenv().getOrDefault("DB_NAME", "inventory_db");

        String jdbcUrl = String.format("jdbc:postgresql://%s:%s/%s", dbHost, dbPort, dbName);
        db = DriverManager.getConnection(jdbcUrl, dbUser, dbPassword);

        // Initialize database
        initDB();

        // HTTP server
        int port = Integer.parseInt(System.getenv().getOrDefault("PORT", "8080"));
        HttpServer server = HttpServer.create(new InetSocketAddress(port), 0);
        server.createContext("/health", InventoryService::healthHandler);
        server.createContext("/api/inventory", InventoryService::inventoryHandler);
        server.setExecutor(null);
        server.start();
        System.out.println("Inventory service starting on port " + port);
    }

    private static void initDB() throws SQLException {
        String createTable = """
            CREATE TABLE IF NOT EXISTS inventory (
                id SERIAL PRIMARY KEY,
                product_id INTEGER NOT NULL,
                warehouse_id INTEGER NOT NULL,
                quantity INTEGER NOT NULL,
                reserved_quantity INTEGER DEFAULT 0,
                last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            );
            """;
        db.createStatement().execute(createTable);

        // Insert sample data
        String insertData = """
            INSERT INTO inventory (product_id, warehouse_id, quantity, reserved_quantity) 
            VALUES 
                (1, 1, 500, 50),
                (2, 1, 300, 20),
                (3, 2, 800, 100),
                (1, 2, 200, 0)
            ON CONFLICT DO NOTHING;
            """;
        try {
            db.createStatement().execute(insertData);
        } catch (SQLException e) {
            // Ignore if data already exists
        }
    }

    private static void healthHandler(HttpExchange exchange) throws IOException {
        try {
            if (db.isValid(5)) {
                sendResponse(exchange, 200, Map.of("status", "healthy"));
            } else {
                sendResponse(exchange, 503, Map.of("status", "unhealthy"));
            }
        } catch (SQLException e) {
            sendResponse(exchange, 503, Map.of("status", "unhealthy", "error", e.getMessage()));
        }
    }

    private static void inventoryHandler(HttpExchange exchange) throws IOException {
        String method = exchange.getRequestMethod();
        if ("GET".equals(method)) {
            getInventory(exchange);
        } else if ("POST".equals(method)) {
            updateInventory(exchange);
        } else {
            sendResponse(exchange, 405, Map.of("error", "Method not allowed"));
        }
    }

    private static void getInventory(HttpExchange exchange) throws IOException {
        try {
            List<Map<String, Object>> items = new ArrayList<>();
            ResultSet rs = db.createStatement().executeQuery(
                "SELECT id, product_id, warehouse_id, quantity, reserved_quantity, last_updated FROM inventory ORDER BY id"
            );
            while (rs.next()) {
                items.add(Map.of(
                    "id", rs.getInt("id"),
                    "product_id", rs.getInt("product_id"),
                    "warehouse_id", rs.getInt("warehouse_id"),
                    "quantity", rs.getInt("quantity"),
                    "reserved_quantity", rs.getInt("reserved_quantity"),
                    "last_updated", rs.getTimestamp("last_updated").toString()
                ));
            }
            sendResponse(exchange, 200, items);
        } catch (SQLException e) {
            sendResponse(exchange, 500, Map.of("error", e.getMessage()));
        }
    }

    private static void updateInventory(HttpExchange exchange) throws IOException {
        try {
            String body = new String(exchange.getRequestBody().readAllBytes());
            Map<String, Object> item = mapper.readValue(body, Map.class);

            String sql = "INSERT INTO inventory (product_id, warehouse_id, quantity, reserved_quantity) VALUES (?, ?, ?, ?) RETURNING id, last_updated";
            PreparedStatement stmt = db.prepareStatement(sql);
            stmt.setInt(1, ((Number) item.get("product_id")).intValue());
            stmt.setInt(2, ((Number) item.get("warehouse_id")).intValue());
            stmt.setInt(3, ((Number) item.get("quantity")).intValue());
            stmt.setInt(4, ((Number) item.getOrDefault("reserved_quantity", 0)).intValue());

            ResultSet rs = stmt.executeQuery();
            if (rs.next()) {
                item.put("id", rs.getInt("id"));
                item.put("last_updated", rs.getTimestamp("last_updated").toString());
                sendResponse(exchange, 201, item);
            }
        } catch (Exception e) {
            sendResponse(exchange, 500, Map.of("error", e.getMessage()));
        }
    }

    private static void sendResponse(HttpExchange exchange, int statusCode, Object response) throws IOException {
        String json = mapper.writeValueAsString(response);
        exchange.getResponseHeaders().set("Content-Type", "application/json");
        exchange.sendResponseHeaders(statusCode, json.getBytes().length);
        OutputStream os = exchange.getResponseBody();
        os.write(json.getBytes());
        os.close();
    }
}

