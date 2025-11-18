package com.sharding.system.client.repository;

import com.sharding.system.client.ShardingClient;
import com.sharding.system.client.ShardingClientException;
import com.sharding.system.client.model.QueryResponse;
import com.sharding.system.client.util.EntityUtils;

import java.lang.reflect.InvocationHandler;
import java.lang.reflect.Method;
import java.lang.reflect.ParameterizedType;
import java.lang.reflect.Type;
import java.util.*;

/**
 * Proxy handler for repository interfaces.
 * Intercepts method calls and executes queries automatically.
 */
public class ShardingRepositoryProxy implements InvocationHandler {
    private final ShardingClient shardingClient;
    @SuppressWarnings("unused")
    private final Class<?> repositoryInterface;
    private final Class<?> entityClass;
    private final Map<Method, QueryMethod> queryMethodCache = new HashMap<>();
    
    public ShardingRepositoryProxy(ShardingClient shardingClient, Class<?> repositoryInterface, Class<?> entityClass) {
        this.shardingClient = shardingClient;
        this.repositoryInterface = repositoryInterface;
        this.entityClass = entityClass;
    }
    
    @Override
    public Object invoke(Object proxy, Method method, Object[] args) throws Throwable {
        // Handle Object methods
        if (method.getDeclaringClass() == Object.class) {
            return method.invoke(this, args);
        }
        
        // Get or create QueryMethod
        QueryMethod queryMethod = queryMethodCache.computeIfAbsent(method, m -> new QueryMethod(m, entityClass));
        
        // Generate query
        String query = queryMethod.generateQuery(args);
        
        // Extract shard key
        String shardKey = queryMethod.extractShardKey(args);
        
        // Extract parameters
        List<Object> params = queryMethod.extractParameters(args);
        
        // Execute query
        QueryResponse response;
        try {
            if (queryMethod.isStrongConsistency()) {
                response = shardingClient.queryStrong(shardKey, query, params);
            } else {
                response = shardingClient.queryEventual(shardKey, query, params);
            }
        } catch (ShardingClientException e) {
            throw new RuntimeException("Query execution failed: " + e.getMessage(), e);
        }
        
        // Map results
        return mapResponse(response, queryMethod.getReturnType(), method);
    }
    
    /**
     * Maps QueryResponse to the method's return type.
     */
    private Object mapResponse(QueryResponse response, Class<?> returnType, Method method) {
        List<Map<String, Object>> rows = response.getRows();
        
        // Handle void
        if (returnType == void.class || returnType == Void.class) {
            return null;
        }
        
        // Handle Optional
        if (returnType == Optional.class) {
            if (rows.isEmpty()) {
                return Optional.empty();
            }
            Type genericReturnType = method.getGenericReturnType();
            if (genericReturnType instanceof ParameterizedType) {
                Type[] actualTypes = ((ParameterizedType) genericReturnType).getActualTypeArguments();
                if (actualTypes.length > 0 && actualTypes[0] instanceof Class) {
                    Class<?> entityType = (Class<?>) actualTypes[0];
                    Object entity = EntityUtils.mapRowToEntity(rows.get(0), entityType);
                    return Optional.of(entity);
                }
            }
            return Optional.of(EntityUtils.mapRowToEntity(rows.get(0), entityClass));
        }
        
        // Handle List
        if (returnType == List.class) {
            Type genericReturnType = method.getGenericReturnType();
            List<Object> result = new ArrayList<>();
            if (genericReturnType instanceof ParameterizedType) {
                Type[] actualTypes = ((ParameterizedType) genericReturnType).getActualTypeArguments();
                if (actualTypes.length > 0 && actualTypes[0] instanceof Class) {
                    Class<?> entityType = (Class<?>) actualTypes[0];
                    for (Map<String, Object> row : rows) {
                        result.add(EntityUtils.mapRowToEntity(row, entityType));
                    }
                    return result;
                }
            }
            // Fallback to entity class
            for (Map<String, Object> row : rows) {
                result.add(EntityUtils.mapRowToEntity(row, entityClass));
            }
            return result;
        }
        
        // Handle boolean (for existsById, etc.)
        if (returnType == boolean.class || returnType == Boolean.class) {
            if (rows.isEmpty()) {
                return false;
            }
            // Check if first row has a count or boolean value
            Map<String, Object> firstRow = rows.get(0);
            if (firstRow.containsKey("count")) {
                Object count = firstRow.get("count");
                if (count instanceof Number) {
                    return ((Number) count).intValue() > 0;
                }
            }
            return !rows.isEmpty();
        }
        
        // Handle long (for count)
        if (returnType == long.class || returnType == Long.class) {
            if (rows.isEmpty()) {
                return 0L;
            }
            Map<String, Object> firstRow = rows.get(0);
            if (firstRow.containsKey("count")) {
                Object count = firstRow.get("count");
                if (count instanceof Number) {
                    return ((Number) count).longValue();
                }
            }
            return (long) rows.size();
        }
        
        // Handle entity type
        if (entityClass.isAssignableFrom(returnType)) {
            if (rows.isEmpty()) {
                return null;
            }
            return EntityUtils.mapRowToEntity(rows.get(0), returnType);
        }
        
        // Handle primitive wrappers
        if (returnType.isPrimitive() || Number.class.isAssignableFrom(returnType)) {
            if (rows.isEmpty()) {
                return null;
            }
            Map<String, Object> firstRow = rows.get(0);
            Collection<Object> values = firstRow.values();
            if (!values.isEmpty()) {
                return values.iterator().next();
            }
        }
        
        // Default: return the entity
        if (rows.isEmpty()) {
            return null;
        }
        return EntityUtils.mapRowToEntity(rows.get(0), entityClass);
    }
}

