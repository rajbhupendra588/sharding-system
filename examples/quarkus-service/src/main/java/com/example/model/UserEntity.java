package com.example.model;

import com.sharding.system.client.annotation.Column;
import com.sharding.system.client.annotation.Entity;
import com.sharding.system.client.annotation.ShardKey;

/**
 * User entity with low-code annotations.
 * No manual mapping needed!
 */
@Entity(table = "users")
public class UserEntity {
    
    @ShardKey
    @Column(name = "id")
    private String id;
    
    @Column(name = "name")
    private String name;
    
    @Column(name = "email")
    private String email;
    
    public UserEntity() {
    }
    
    public UserEntity(String id, String name, String email) {
        this.id = id;
        this.name = name;
        this.email = email;
    }
    
    public String getId() {
        return id;
    }
    
    public void setId(String id) {
        this.id = id;
    }
    
    public String getName() {
        return name;
    }
    
    public void setName(String name) {
        this.name = name;
    }
    
    public String getEmail() {
        return email;
    }
    
    public void setEmail(String email) {
        this.email = email;
    }
}

