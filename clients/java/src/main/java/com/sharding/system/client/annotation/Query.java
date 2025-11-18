package com.sharding.system.client.annotation;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Provides a custom SQL query for a repository method.
 * 
 * Example:
 * <pre>
 * {@code
 * @Query("SELECT * FROM users WHERE email = $1 AND status = $2")
 * List<User> findByEmailAndStatus(String email, String status);
 * }
 * </pre>
 */
@Target(ElementType.METHOD)
@Retention(RetentionPolicy.RUNTIME)
public @interface Query {
    /**
     * The SQL query string. Use $1, $2, etc. for parameters.
     */
    String value();
}

