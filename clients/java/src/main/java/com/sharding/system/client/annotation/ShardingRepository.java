package com.sharding.system.client.annotation;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Marks an interface as a sharding repository.
 * The framework will automatically generate implementations for all methods.
 * 
 * Example:
 * <pre>
 * {@code
 * @ShardingRepository(entity = User.class, table = "users")
 * public interface UserRepository extends CrudRepository<User, String> {
 *     Optional<User> findByEmail(String email);
 * }
 * }
 * </pre>
 */
@Target(ElementType.TYPE)
@Retention(RetentionPolicy.RUNTIME)
public @interface ShardingRepository {
    /**
     * The entity class this repository manages.
     */
    Class<?> entity();
    
    /**
     * The table name. If not specified, will be derived from entity @Entity annotation.
     */
    String table() default "";
}

