package com.example.service;

import com.example.model.UserEntity;
import com.example.repository.UserRepository;
import com.sharding.system.client.config.ShardingClientAutoConfiguration;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import jakarta.enterprise.context.ApplicationScoped;
import org.eclipse.microprofile.config.inject.ConfigProperty;
import org.jboss.logging.Logger;

import java.util.List;
import java.util.Optional;

/**
 * User service using LOW-CODE approach.
 * 90-99% less code compared to manual implementation!
 */
@ApplicationScoped
public class LowCodeUserService {
    
    private static final Logger LOG = Logger.getLogger(LowCodeUserService.class);
    
    @ConfigProperty(name = "sharding.router.url", defaultValue = "http://localhost:8080")
    String routerUrl;
    
    private ShardingClientAutoConfiguration config;
    private UserRepository userRepository;
    
    @PostConstruct
    public void init() {
        // Initialize auto-configuration
        config = new ShardingClientAutoConfiguration();
        config.setRouterUrl(routerUrl);
        config.initialize();
        
        // Get repository - that's it!
        userRepository = config.getRepository(UserRepository.class);
        
        LOG.info("Low-code user service initialized");
    }
    
    @PreDestroy
    public void destroy() {
        if (config != null) {
            config.close();
        }
    }
    
    /**
     * Get user by ID - ONE LINE!
     */
    public Optional<UserEntity> getUserById(String userId) {
        return userRepository.findById(userId);
    }
    
    /**
     * Create user - ONE LINE!
     */
    public UserEntity createUser(UserEntity user) {
        return userRepository.save(user);
    }
    
    /**
     * Update user - ONE LINE!
     */
    public UserEntity updateUser(UserEntity user) {
        return userRepository.save(user);
    }
    
    /**
     * Delete user - ONE LINE!
     */
    public void deleteUser(String userId) {
        userRepository.deleteById(userId);
    }
    
    /**
     * Find by email - AUTO-GENERATED QUERY!
     */
    public Optional<UserEntity> findByEmail(String email) {
        return userRepository.findByEmail(email);
    }
    
    /**
     * Find by name pattern - CUSTOM QUERY WITH AUTO-MAPPING!
     */
    public List<UserEntity> findByNameLike(String pattern, int limit) {
        return userRepository.findByNameLike(pattern, limit);
    }
    
    /**
     * Find by status - EVENTUAL CONSISTENCY!
     */
    public List<UserEntity> findByStatus(String status) {
        return userRepository.findByStatus(status);
    }
    
    /**
     * List all users - ONE LINE!
     */
    public List<UserEntity> listAllUsers(String shardKey) {
        return userRepository.findAll(shardKey);
    }
    
    /**
     * Check if user exists - ONE LINE!
     */
    public boolean userExists(String userId) {
        return userRepository.existsById(userId);
    }
}

