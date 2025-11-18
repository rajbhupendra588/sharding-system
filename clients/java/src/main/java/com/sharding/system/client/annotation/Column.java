package com.sharding.system.client.annotation;

import java.lang.annotation.ElementType;
import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;

/**
 * Maps a field to a database column.
 * If not specified, defaults to the field name in snake_case.
 */
@Target(ElementType.FIELD)
@Retention(RetentionPolicy.RUNTIME)
public @interface Column {
    /**
     * The column name. If not specified, defaults to the field name in snake_case.
     */
    String name() default "";
    
    /**
     * Whether this column is nullable.
     */
    boolean nullable() default true;
}

