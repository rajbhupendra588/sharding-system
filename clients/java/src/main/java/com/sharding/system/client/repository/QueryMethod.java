package com.sharding.system.client.repository;

import com.sharding.system.client.annotation.*;
import com.sharding.system.client.util.EntityUtils;

import java.lang.reflect.Field;
import java.lang.reflect.Method;
import java.lang.reflect.Parameter;
import java.util.*;

/**
 * Represents a query method in a repository.
 * Parses method names and annotations to generate SQL queries.
 */
public class QueryMethod {
    private final Method method;
    private final Class<?> entityClass;
    private final String tableName;
    private final boolean hasCustomQuery;
    private final String customQuery;
    private final boolean isStrongConsistency;
    private final String methodName;
    private final Class<?> returnType;
    private final Parameter[] parameters;
    
    public QueryMethod(Method method, Class<?> entityClass) {
        this.method = method;
        this.entityClass = entityClass;
        this.tableName = EntityUtils.getTableName(entityClass);
        this.methodName = method.getName();
        this.returnType = method.getReturnType();
        this.parameters = method.getParameters();
        
        // Check for custom query
        Query queryAnnotation = method.getAnnotation(Query.class);
        this.hasCustomQuery = queryAnnotation != null;
        this.customQuery = hasCustomQuery ? queryAnnotation.value() : null;
        
        // Check consistency annotations
        boolean hasStrong = method.isAnnotationPresent(StrongConsistency.class);
        boolean hasEventual = method.isAnnotationPresent(EventualConsistency.class);
        this.isStrongConsistency = hasStrong || (!hasEventual && isWriteOperation());
    }
    
    /**
     * Determines if this is a write operation (INSERT, UPDATE, DELETE).
     */
    private boolean isWriteOperation() {
        String name = methodName.toLowerCase();
        return name.startsWith("save") || 
               name.startsWith("insert") || 
               name.startsWith("update") || 
               name.startsWith("delete") || 
               name.startsWith("remove");
    }
    
    /**
     * Generates SQL query for this method.
     */
    public String generateQuery(Object[] args) {
        if (hasCustomQuery) {
            return customQuery;
        }
        
        // Handle CRUD operations
        if (methodName.equals("save") || methodName.equals("saveAll")) {
            return EntityUtils.buildInsertQuery(entityClass);
        }
        
        if (methodName.equals("findById")) {
            Field shardKeyField = EntityUtils.findShardKeyField(entityClass);
            String idColumn = EntityUtils.getColumnName(shardKeyField);
            return EntityUtils.buildSelectQuery(entityClass, idColumn + " = $1");
        }
        
        if (methodName.equals("existsById")) {
            Field shardKeyField = EntityUtils.findShardKeyField(entityClass);
            String idColumn = EntityUtils.getColumnName(shardKeyField);
            return "SELECT COUNT(*) > 0 FROM " + tableName + " WHERE " + idColumn + " = $1";
        }
        
        if (methodName.equals("findAll")) {
            return EntityUtils.buildSelectQuery(entityClass, null);
        }
        
        if (methodName.equals("count")) {
            return "SELECT COUNT(*) FROM " + tableName;
        }
        
        if (methodName.equals("deleteById") || methodName.equals("delete")) {
            Field shardKeyField = EntityUtils.findShardKeyField(entityClass);
            String idColumn = EntityUtils.getColumnName(shardKeyField);
            return "DELETE FROM " + tableName + " WHERE " + idColumn + " = $1";
        }
        
        if (methodName.equals("deleteAll")) {
            return "DELETE FROM " + tableName;
        }
        
        // Parse method name for findBy* patterns
        if (methodName.startsWith("findBy")) {
            return parseFindByMethod(args);
        }
        
        throw new UnsupportedOperationException("Cannot generate query for method: " + methodName);
    }
    
    /**
     * Parses findBy* method names (e.g., findByEmail, findByEmailAndStatus).
     */
    private String parseFindByMethod(Object[] args) {
        String queryPart = methodName.substring(6); // Remove "findBy"
        String[] parts = splitCamelCase(queryPart);
        
        List<String> conditions = new ArrayList<>();
        int paramIndex = 1;
        
        for (int i = 0; i < parts.length; i += 2) {
            if (i + 1 >= parts.length) {
                // Single field
                String fieldName = camelToSnake(parts[i]);
                conditions.add(fieldName + " = $" + paramIndex++);
                break;
            }
            
            String fieldName = camelToSnake(parts[i]);
            String operator = parts[i + 1].toLowerCase();
            
            if (operator.equals("and")) {
                conditions.add(fieldName + " = $" + paramIndex++);
            } else if (operator.equals("or")) {
                // Handle OR conditions
                conditions.add(fieldName + " = $" + paramIndex++);
            } else {
                // Single field
                conditions.add(fieldName + " = $" + paramIndex++);
                i--; // Adjust index
            }
        }
        
        String whereClause = String.join(" AND ", conditions);
        return EntityUtils.buildSelectQuery(entityClass, whereClause);
    }
    
    /**
     * Splits camelCase into parts.
     */
    private String[] splitCamelCase(String str) {
        List<String> parts = new ArrayList<>();
        StringBuilder current = new StringBuilder();
        
        for (char c : str.toCharArray()) {
            if (Character.isUpperCase(c) && current.length() > 0) {
                parts.add(current.toString());
                current = new StringBuilder();
            }
            current.append(c);
        }
        if (current.length() > 0) {
            parts.add(current.toString());
        }
        
        return parts.toArray(new String[0]);
    }
    
    /**
     * Converts camelCase to snake_case.
     */
    private String camelToSnake(String str) {
        if (str == null || str.isEmpty()) {
            return str;
        }
        StringBuilder result = new StringBuilder();
        result.append(Character.toLowerCase(str.charAt(0)));
        for (int i = 1; i < str.length(); i++) {
            char c = str.charAt(i);
            if (Character.isUpperCase(c)) {
                result.append('_').append(Character.toLowerCase(c));
            } else {
                result.append(c);
            }
        }
        return result.toString();
    }
    
    /**
     * Extracts shard key from method arguments.
     */
    public String extractShardKey(Object[] args) {
        // Check parameters for @ShardKey annotation
        for (int i = 0; i < parameters.length; i++) {
            if (parameters[i].isAnnotationPresent(ShardKey.class) && args[i] != null) {
                return args[i].toString();
            }
        }
        
        // Check first argument if it's an entity
        if (args.length > 0 && args[0] != null) {
            Object firstArg = args[0];
            if (firstArg.getClass().isAnnotationPresent(Entity.class)) {
                return EntityUtils.extractShardKey(firstArg);
            }
            // For findById, the ID is the shard key
            if (methodName.equals("findById") || methodName.equals("deleteById") || methodName.equals("existsById")) {
                return firstArg.toString();
            }
        }
        
        // For findAll, count, deleteAll - check for shardKey parameter
        for (int i = 0; i < parameters.length; i++) {
            if (parameters[i].getName().equals("shardKey") && args[i] != null) {
                return args[i].toString();
            }
        }
        
        throw new IllegalArgumentException("Cannot determine shard key for method: " + methodName);
    }
    
    /**
     * Extracts query parameters from method arguments.
     */
    public List<Object> extractParameters(Object[] args) {
        List<Object> params = new ArrayList<>();
        
        if (hasCustomQuery) {
            // For custom queries, use all arguments as parameters
            params.addAll(Arrays.asList(args));
            return params;
        }
        
        // Handle CRUD operations
        if (methodName.equals("save")) {
            return EntityUtils.extractInsertValues(args[0]);
        }
        
        if (methodName.equals("findById") || methodName.equals("deleteById") || methodName.equals("existsById")) {
            params.add(args[0]);
            return params;
        }
        
        if (methodName.equals("findAll") || methodName.equals("count") || methodName.equals("deleteAll")) {
            // These methods might have a shardKey parameter
            if (args.length > 0) {
                params.add(args[0]); // shardKey
            }
            return params;
        }
        
        // For findBy* methods, use all arguments
        params.addAll(Arrays.asList(args));
        return params;
    }
    
    public boolean isStrongConsistency() {
        return isStrongConsistency;
    }
    
    public Class<?> getReturnType() {
        return returnType;
    }
    
    public Method getMethod() {
        return method;
    }
}

