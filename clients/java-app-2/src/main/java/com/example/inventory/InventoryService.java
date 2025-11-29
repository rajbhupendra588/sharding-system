package com.example.inventory;

import com.google.gson.Gson;
import spark.Request;
import spark.Response;
import spark.Spark;

import java.sql.*;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class InventoryService {
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
        String dbName = System.getenv().getOrDefault("DB_NAME", "inventory_db");

        String url = String.format("jdbc:postgresql://%s:%s/%s", dbHost, dbPort, dbName);

        try {
            db = DriverManager.getConnection(url, dbUser, dbPassword);
            System.out.println("Connected to database successfully");

            String createTable = """
                CREATE TABLE IF NOT EXISTS inventory_items (
                    id SERIAL PRIMARY KEY,
                    product_id INTEGER NOT NULL,
                    warehouse_location VARCHAR(100) NOT NULL,
                    quantity INTEGER NOT NULL DEFAULT 0,
                    reserved_quantity INTEGER DEFAULT 0,
                    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                );
                """;

            try (Statement stmt = db.createStatement()) {
                stmt.execute(createTable);
            }

            String insertData = """
                INSERT INTO inventory_items (product_id, warehouse_location, quantity, reserved_quantity)
                VALUES 
                    (1, 'Warehouse A', 100, 10),
                    (2, 'Warehouse B', 250, 25),
                    (3, 'Warehouse A', 150, 5)
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

        Spark.get("/health", InventoryService::healthCheck);
        Spark.get("/inventory", InventoryService::getInventory);

        System.out.println("Inventory service started on port " + System.getenv().getOrDefault("PORT", "8080"));
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

    private static String getInventory(Request req, Response res) {
        List<Map<String, Object>> items = new ArrayList<>();
        try {
            String query = "SELECT id, product_id, warehouse_location, quantity, reserved_quantity, last_updated FROM inventory_items";
            try (Statement stmt = db.createStatement();
                 ResultSet rs = stmt.executeQuery(query)) {
                while (rs.next()) {
                    Map<String, Object> item = new HashMap<>();
                    item.put("id", rs.getInt("id"));
                    item.put("product_id", rs.getInt("product_id"));
                    item.put("warehouse_location", rs.getString("warehouse_location"));
                    item.put("quantity", rs.getInt("quantity"));
                    item.put("reserved_quantity", rs.getInt("reserved_quantity"));
                    item.put("last_updated", rs.getTimestamp("last_updated").toString());
                    items.add(item);
                }
            }
            res.type("application/json");
            return gson.toJson(Map.of("inventory", items));
        } catch (SQLException e) {
            res.status(500);
            return "{\"error\": \"" + e.getMessage() + "\"}";
        }
    }
}

