package com.sharding;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.sun.net.httpserver.HttpExchange;
import com.sun.net.httpserver.HttpServer;

import java.io.IOException;
import java.io.OutputStream;
import java.net.InetSocketAddress;
import java.sql.*;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;

public class ProductService {
    private static Connection db;
    private static ObjectMapper mapper = new ObjectMapper();

    public static void main(String[] args) throws Exception {
        // Database connection
        String dbHost = System.getenv().getOrDefault("DB_HOST", "localhost");
        String dbPort = System.getenv().getOrDefault("DB_PORT", "5432");
        String dbUser = System.getenv().getOrDefault("DB_USER", "postgres");
        String dbPassword = System.getenv().getOrDefault("DB_PASSWORD", "postgres");
        String dbName = System.getenv().getOrDefault("DB_NAME", "products_db");

        String jdbcUrl = String.format("jdbc:postgresql://%s:%s/%s", dbHost, dbPort, dbName);
        db = DriverManager.getConnection(jdbcUrl, dbUser, dbPassword);

        // Initialize database
        initDB();

        // HTTP server
        int port = Integer.parseInt(System.getenv().getOrDefault("PORT", "8080"));
        HttpServer server = HttpServer.create(new InetSocketAddress(port), 0);
        server.createContext("/health", ProductService::healthHandler);
        server.createContext("/api/products", ProductService::productsHandler);
        server.setExecutor(null);
        server.start();
        System.out.println("Product service starting on port " + port);
    }

    private static void initDB() throws SQLException {
        String createTable = """
            CREATE TABLE IF NOT EXISTS products (
                id SERIAL PRIMARY KEY,
                name VARCHAR(255) NOT NULL,
                category VARCHAR(100),
                price DECIMAL(10,2) NOT NULL,
                stock INTEGER DEFAULT 0,
                created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
            );
            """;
        db.createStatement().execute(createTable);

        // Insert sample data
        String insertData = """
            INSERT INTO products (name, category, price, stock) 
            VALUES 
                ('Widget A', 'Electronics', 29.99, 100),
                ('Widget B', 'Electronics', 49.99, 50),
                ('Widget C', 'Accessories', 19.99, 200)
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

    private static void productsHandler(HttpExchange exchange) throws IOException {
        String method = exchange.getRequestMethod();
        if ("GET".equals(method)) {
            getProducts(exchange);
        } else if ("POST".equals(method)) {
            createProduct(exchange);
        } else {
            sendResponse(exchange, 405, Map.of("error", "Method not allowed"));
        }
    }

    private static void getProducts(HttpExchange exchange) throws IOException {
        try {
            List<Map<String, Object>> products = new ArrayList<>();
            ResultSet rs = db.createStatement().executeQuery(
                "SELECT id, name, category, price, stock, created_at FROM products ORDER BY id"
            );
            while (rs.next()) {
                products.add(Map.of(
                    "id", rs.getInt("id"),
                    "name", rs.getString("name"),
                    "category", rs.getString("category"),
                    "price", rs.getDouble("price"),
                    "stock", rs.getInt("stock"),
                    "created_at", rs.getTimestamp("created_at").toString()
                ));
            }
            sendResponse(exchange, 200, products);
        } catch (SQLException e) {
            sendResponse(exchange, 500, Map.of("error", e.getMessage()));
        }
    }

    private static void createProduct(HttpExchange exchange) throws IOException {
        try {
            String body = new String(exchange.getRequestBody().readAllBytes());
            Map<String, Object> product = mapper.readValue(body, Map.class);

            String sql = "INSERT INTO products (name, category, price, stock) VALUES (?, ?, ?, ?) RETURNING id, created_at";
            PreparedStatement stmt = db.prepareStatement(sql);
            stmt.setString(1, (String) product.get("name"));
            stmt.setString(2, (String) product.getOrDefault("category", ""));
            stmt.setDouble(3, ((Number) product.get("price")).doubleValue());
            stmt.setInt(4, ((Number) product.getOrDefault("stock", 0)).intValue());

            ResultSet rs = stmt.executeQuery();
            if (rs.next()) {
                product.put("id", rs.getInt("id"));
                product.put("created_at", rs.getTimestamp("created_at").toString());
                sendResponse(exchange, 201, product);
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

