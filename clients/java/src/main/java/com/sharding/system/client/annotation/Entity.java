package com.sharding.system.client.annotation;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Marks a class as a sharding entity.
 * Entities are automatically mapped to database tables.
 */
@Target(ElementType.TYPE)
@Retention(RetentionPolicy.RUNTIME)
public @interface Entity {
    /**
     * The table name. If not specified, defaults to the class name in snake_case.
     */
    String table() default "";
}

