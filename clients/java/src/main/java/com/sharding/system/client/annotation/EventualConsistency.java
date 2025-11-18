package com.sharding.system.client.annotation;

import java.lang.annotation.Retention;
import java.lang.annotation.RetentionPolicy;
import java.lang.annotation.Target;
import java.lang.annotation.ElementType;

/**
 * Marks a method to use eventual consistency (can read from replica).
 * Useful for read-heavy operations that don't require the latest data.
 */
@Target(ElementType.METHOD)
@Retention(RetentionPolicy.RUNTIME)
public @interface EventualConsistency {
}

