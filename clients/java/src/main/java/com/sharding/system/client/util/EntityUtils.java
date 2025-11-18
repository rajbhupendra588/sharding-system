package com.sharding.system.client.util;

import com.sharding.system.client.annotation.Column;
import com.sharding.system.client.annotation.Entity;
import com.sharding.system.client.annotation.ShardKey;

import java.lang.reflect.Field;
import java.util.*;

/**
 * Utility class for entity operations.
 */
public class EntityUtils {
    
    /**
     * Gets the table name for an entity class.
     */
    public static String getTableName(Class<?> entityClass) {
        Entity entityAnnotation = entityClass.getAnnotation(Entity.class);
        if (entityAnnotation != null && !entityAnnotation.table().isEmpty()) {
            return entityAnnotation.table();
        }
        // Convert class name to snake_case
        return camelToSnake(entityClass.getSimpleName());
    }
    
    /**
     * Finds the shard key field in an entity.
     */
    public static Field findShardKeyField(Class<?> entityClass) {
        for (Field field : getAllFields(entityClass)) {
            if (field.isAnnotationPresent(ShardKey.class)) {
                field.setAccessible(true);
                return field;
            }
        }
        // Default: look for "id" field
        try {
            Field idField = entityClass.getDeclaredField("id");
            idField.setAccessible(true);
            return idField;
        } catch (NoSuchFieldException e) {
            throw new IllegalArgumentException("No shard key field found in entity: " + entityClass.getName());
        }
    }
    
    /**
     * Extracts the shard key value from an entity.
     */
    public static String extractShardKey(Object entity) {
        if (entity == null) {
            throw new IllegalArgumentException("Entity cannot be null");
        }
        Field shardKeyField = findShardKeyField(entity.getClass());
        try {
            Object value = shardKeyField.get(entity);
            return value != null ? value.toString() : null;
        } catch (IllegalAccessException e) {
            throw new RuntimeException("Failed to extract shard key from entity", e);
        }
    }
    
    /**
     * Extracts the shard key value from a parameter.
     */
    public static String extractShardKeyFromParam(Object param, java.lang.reflect.Parameter parameter) {
        if (param == null) {
            return null;
        }
        if (parameter.isAnnotationPresent(ShardKey.class)) {
            return param.toString();
        }
        // If param is an entity, extract shard key from it
        if (param.getClass().isAnnotationPresent(Entity.class)) {
            return extractShardKey(param);
        }
        return null;
    }
    
    /**
     * Gets all fields including inherited ones.
     */
    private static List<Field> getAllFields(Class<?> clazz) {
        List<Field> fields = new ArrayList<>();
        while (clazz != null && clazz != Object.class) {
            fields.addAll(Arrays.asList(clazz.getDeclaredFields()));
            clazz = clazz.getSuperclass();
        }
        return fields;
    }
    
    /**
     * Gets the column name for a field.
     */
    public static String getColumnName(Field field) {
        Column columnAnnotation = field.getAnnotation(Column.class);
        if (columnAnnotation != null && !columnAnnotation.name().isEmpty()) {
            return columnAnnotation.name();
        }
        return camelToSnake(field.getName());
    }
    
    /**
     * Converts camelCase to snake_case.
     */
    private static String camelToSnake(String str) {
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
     * Maps a row (Map<String, Object>) to an entity.
     */
    public static <T> T mapRowToEntity(Map<String, Object> row, Class<T> entityClass) {
        try {
            T entity = entityClass.getDeclaredConstructor().newInstance();
            List<Field> fields = getAllFields(entityClass);
            
            for (Field field : fields) {
                field.setAccessible(true);
                String columnName = getColumnName(field);
                Object value = row.get(columnName);
                
                if (value != null) {
                    // Type conversion
                    Object convertedValue = convertValue(value, field.getType());
                    field.set(entity, convertedValue);
                }
            }
            
            return entity;
        } catch (Exception e) {
            throw new RuntimeException("Failed to map row to entity: " + entityClass.getName(), e);
        }
    }
    
    /**
     * Converts a value to the target type.
     */
    private static Object convertValue(Object value, Class<?> targetType) {
        if (value == null) {
            return null;
        }
        
        if (targetType.isAssignableFrom(value.getClass())) {
            return value;
        }
        
        // String conversions
        if (targetType == String.class) {
            return value.toString();
        }
        
        // Number conversions
        if (targetType == Integer.class || targetType == int.class) {
            if (value instanceof Number) {
                return ((Number) value).intValue();
            }
            return Integer.parseInt(value.toString());
        }
        
        if (targetType == Long.class || targetType == long.class) {
            if (value instanceof Number) {
                return ((Number) value).longValue();
            }
            return Long.parseLong(value.toString());
        }
        
        if (targetType == Double.class || targetType == double.class) {
            if (value instanceof Number) {
                return ((Number) value).doubleValue();
            }
            return Double.parseDouble(value.toString());
        }
        
        if (targetType == Boolean.class || targetType == boolean.class) {
            if (value instanceof Boolean) {
                return value;
            }
            return Boolean.parseBoolean(value.toString());
        }
        
        return value;
    }
    
    /**
     * Builds an INSERT query for an entity.
     */
    public static String buildInsertQuery(Class<?> entityClass) {
        String tableName = getTableName(entityClass);
        List<Field> fields = getAllFields(entityClass);
        
        List<String> columns = new ArrayList<>();
        List<String> placeholders = new ArrayList<>();
        
        int paramIndex = 1;
        for (Field field : fields) {
            columns.add(getColumnName(field));
            placeholders.add("$" + paramIndex++);
        }
        
        return String.format("INSERT INTO %s (%s) VALUES (%s)",
            tableName,
            String.join(", ", columns),
            String.join(", ", placeholders));
    }
    
    /**
     * Builds an UPDATE query for an entity.
     */
    public static String buildUpdateQuery(Class<?> entityClass) {
        String tableName = getTableName(entityClass);
        Field shardKeyField = findShardKeyField(entityClass);
        String idColumn = getColumnName(shardKeyField);
        List<Field> fields = getAllFields(entityClass);
        
        List<String> setClauses = new ArrayList<>();
        int paramIndex = 1;
        
        for (Field field : fields) {
            if (field.equals(shardKeyField)) {
                continue; // Skip ID in SET clause
            }
            setClauses.add(getColumnName(field) + " = $" + paramIndex++);
        }
        
        // Add ID as last parameter
        setClauses.add(idColumn + " = $" + paramIndex);
        
        return String.format("UPDATE %s SET %s WHERE %s = $%d",
            tableName,
            String.join(", ", setClauses.subList(0, setClauses.size() - 1)),
            idColumn,
            paramIndex);
    }
    
    /**
     * Builds a SELECT query for an entity.
     */
    public static String buildSelectQuery(Class<?> entityClass, String whereClause) {
        String tableName = getTableName(entityClass);
        List<Field> fields = getAllFields(entityClass);
        List<String> columns = new ArrayList<>();
        
        for (Field field : fields) {
            columns.add(getColumnName(field));
        }
        
        String query = String.format("SELECT %s FROM %s", String.join(", ", columns), tableName);
        if (whereClause != null && !whereClause.isEmpty()) {
            query += " WHERE " + whereClause;
        }
        
        return query;
    }
    
    /**
     * Extracts parameter values from an entity for INSERT.
     */
    public static List<Object> extractInsertValues(Object entity) {
        List<Object> values = new ArrayList<>();
        List<Field> fields = getAllFields(entity.getClass());
        
        for (Field field : fields) {
            field.setAccessible(true);
            try {
                values.add(field.get(entity));
            } catch (IllegalAccessException e) {
                throw new RuntimeException("Failed to extract field value", e);
            }
        }
        
        return values;
    }
    
    /**
     * Extracts parameter values from an entity for UPDATE.
     */
    public static List<Object> extractUpdateValues(Object entity) {
        List<Object> values = new ArrayList<>();
        Field shardKeyField = findShardKeyField(entity.getClass());
        List<Field> fields = getAllFields(entity.getClass());
        
        for (Field field : fields) {
            if (field.equals(shardKeyField)) {
                continue; // Skip ID, will be added last
            }
            field.setAccessible(true);
            try {
                values.add(field.get(entity));
            } catch (IllegalAccessException e) {
                throw new RuntimeException("Failed to extract field value", e);
            }
        }
        
        // Add ID as last parameter
        shardKeyField.setAccessible(true);
        try {
            values.add(shardKeyField.get(entity));
        } catch (IllegalAccessException e) {
            throw new RuntimeException("Failed to extract shard key value", e);
        }
        
        return values;
    }
}

