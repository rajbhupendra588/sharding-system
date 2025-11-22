package com.example.resource;

import com.example.model.User;
import com.example.service.UserService;
import com.sharding.system.client.ShardingClientException;
import jakarta.inject.Inject;
import jakarta.ws.rs.*;
import jakarta.ws.rs.core.MediaType;
import jakarta.ws.rs.core.Response;
import org.jboss.logging.Logger;

import java.util.List;

@Path("/users")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class UserResource {
    
    private static final Logger LOG = Logger.getLogger(UserResource.class);
    
    @Inject
    UserService userService;
    
    @GET
    @Path("/{id}")
    public Response getUser(@PathParam("id") String id) {
        try {
            User user = userService.getUserById(id);
            if (user == null) {
                return Response.status(Response.Status.NOT_FOUND)
                    .entity(new ErrorResponse("User not found: " + id))
                    .build();
            }
            return Response.ok(user).build();
        } catch (ShardingClientException e) {
            LOG.errorf(e, "Failed to get user: %s", id);
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .entity(new ErrorResponse("Failed to get user: " + e.getMessage()))
                .build();
        }
    }
    
    @POST
    public Response createUser(User user) {
        try {
            userService.createUser(user);
            return Response.status(Response.Status.CREATED).build();
        } catch (ShardingClientException e) {
            LOG.errorf(e, "Failed to create user: %s", user.getId());
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .entity(new ErrorResponse("Failed to create user: " + e.getMessage()))
                .build();
        }
    }
    
    @PUT
    @Path("/{id}")
    public Response updateUser(@PathParam("id") String id, User user) {
        try {
            user.setId(id);
            userService.updateUser(user);
            return Response.ok().build();
        } catch (ShardingClientException e) {
            LOG.errorf(e, "Failed to update user: %s", id);
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .entity(new ErrorResponse("Failed to update user: " + e.getMessage()))
                .build();
        }
    }
    
    @DELETE
    @Path("/{id}")
    public Response deleteUser(@PathParam("id") String id) {
        try {
            userService.deleteUser(id);
            return Response.noContent().build();
        } catch (ShardingClientException e) {
            LOG.errorf(e, "Failed to delete user: %s", id);
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .entity(new ErrorResponse("Failed to delete user: " + e.getMessage()))
                .build();
        }
    }
    
    @GET
    public Response listUsers(@QueryParam("shardKey") String shardKey) {
        try {
            String key = shardKey != null ? shardKey : "default";
            List<User> users = userService.listUsers(key);
            return Response.ok(users).build();
        } catch (ShardingClientException e) {
            LOG.errorf(e, "Failed to list users");
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .entity(new ErrorResponse("Failed to list users: " + e.getMessage()))
                .build();
        }
    }
    
    @GET
    @Path("/shard/{key}")
    public Response getShardForKey(@PathParam("key") String key) {
        try {
            String shardId = userService.getShardForKey(key);
            return Response.ok(new ShardInfoResponse(shardId)).build();
        } catch (ShardingClientException e) {
            LOG.errorf(e, "Failed to get shard for key: %s", key);
            return Response.status(Response.Status.INTERNAL_SERVER_ERROR)
                .entity(new ErrorResponse("Failed to get shard: " + e.getMessage()))
                .build();
        }
    }
    
    // Inner classes for responses
    public static class ErrorResponse {
        public String error;
        
        public ErrorResponse(String error) {
            this.error = error;
        }
    }
    
    public static class ShardInfoResponse {
        public String shardId;
        
        public ShardInfoResponse(String shardId) {
            this.shardId = shardId;
        }
    }
}


