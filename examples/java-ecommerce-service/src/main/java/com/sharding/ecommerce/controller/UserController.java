package com.sharding.ecommerce.controller;

import com.sharding.ecommerce.model.User;
import com.sharding.ecommerce.service.UserService;
import com.sharding.system.client.ShardingClientException;
import io.swagger.v3.oas.annotations.Operation;
import io.swagger.v3.oas.annotations.Parameter;
import io.swagger.v3.oas.annotations.tags.Tag;
import jakarta.validation.Valid;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

/**
 * REST Controller for User operations.
 * 
 * Demonstrates how sharding by user ID enables:
 * - Fast user lookups (single shard query)
 * - Efficient user order history (co-located data)
 * - Horizontal scaling as user base grows
 */
@RestController
@RequestMapping("/api/v1/users")
@Tag(name = "Users", description = "User management API")
public class UserController {

    private static final Logger log = LoggerFactory.getLogger(UserController.class);
    private final UserService userService;

    public UserController(UserService userService) {
        this.userService = userService;
    }

    @PostMapping
    @Operation(summary = "Create a new user", description = "Creates a new user. User ID is used as shard key.")
    public ResponseEntity<User> createUser(@Valid @RequestBody User user) {
        try {
            User created = userService.createUser(user);
            return ResponseEntity.status(HttpStatus.CREATED).body(created);
        } catch (ShardingClientException e) {
            log.error("Error creating user", e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @GetMapping("/{id}")
    @Operation(summary = "Get user by ID", description = "Retrieves user information. Uses user ID as shard key for routing.")
    public ResponseEntity<User> getUserById(
            @Parameter(description = "User ID", required = true) @PathVariable String id) {
        try {
            User user = userService.getUserById(id);
            if (user == null) {
                return ResponseEntity.notFound().build();
            }
            return ResponseEntity.ok(user);
        } catch (ShardingClientException e) {
            log.error("Error fetching user: {}", id, e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @PutMapping("/{id}")
    @Operation(summary = "Update user", description = "Updates user information. Uses user ID as shard key.")
    public ResponseEntity<User> updateUser(
            @Parameter(description = "User ID", required = true) @PathVariable String id,
            @Valid @RequestBody User user) {
        try {
            User updated = userService.updateUser(id, user);
            return ResponseEntity.ok(updated);
        } catch (ShardingClientException e) {
            log.error("Error updating user: {}", id, e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @DeleteMapping("/{id}")
    @Operation(summary = "Delete user", description = "Soft deletes a user by setting active=false.")
    public ResponseEntity<Void> deleteUser(
            @Parameter(description = "User ID", required = true) @PathVariable String id) {
        try {
            userService.deleteUser(id);
            return ResponseEntity.noContent().build();
        } catch (ShardingClientException e) {
            log.error("Error deleting user: {}", id, e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }

    @GetMapping("/{id}/shard")
    @Operation(summary = "Get shard for user", 
               description = "Returns which shard contains this user's data. Useful for debugging and monitoring.")
    public ResponseEntity<Map<String, String>> getShardForUser(
            @Parameter(description = "User ID", required = true) @PathVariable String id) {
        try {
            String shardId = userService.getShardForUser(id);
            return ResponseEntity.ok(Map.of("user_id", id, "shard_id", shardId));
        } catch (ShardingClientException e) {
            log.error("Error getting shard for user: {}", id, e);
            return ResponseEntity.status(HttpStatus.INTERNAL_SERVER_ERROR).build();
        }
    }
}

