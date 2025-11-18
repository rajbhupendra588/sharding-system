package com.example.resource;

import com.example.model.UserEntity;
import com.example.service.LowCodeUserService;
import jakarta.inject.Inject;
import jakarta.ws.rs.*;
import jakarta.ws.rs.core.MediaType;
import jakarta.ws.rs.core.Response;

import java.util.List;
import java.util.Optional;

/**
 * REST resource using LOW-CODE approach.
 * Minimal code, maximum functionality!
 */
@Path("/lowcode/users")
@Produces(MediaType.APPLICATION_JSON)
@Consumes(MediaType.APPLICATION_JSON)
public class LowCodeUserResource {
    
    @Inject
    LowCodeUserService userService;
    
    @GET
    @Path("/{id}")
    public Response getUser(@PathParam("id") String id) {
        Optional<UserEntity> user = userService.getUserById(id);
        if (user.isPresent()) {
            return Response.ok(user.get()).build();
        }
        return Response.status(Response.Status.NOT_FOUND).build();
    }
    
    @POST
    public Response createUser(UserEntity user) {
        UserEntity created = userService.createUser(user);
        return Response.status(Response.Status.CREATED).entity(created).build();
    }
    
    @PUT
    @Path("/{id}")
    public Response updateUser(@PathParam("id") String id, UserEntity user) {
        user.setId(id);
        UserEntity updated = userService.updateUser(user);
        return Response.ok(updated).build();
    }
    
    @DELETE
    @Path("/{id}")
    public Response deleteUser(@PathParam("id") String id) {
        userService.deleteUser(id);
        return Response.noContent().build();
    }
    
    @GET
    @Path("/email/{email}")
    public Response getUserByEmail(@PathParam("email") String email) {
        Optional<UserEntity> user = userService.findByEmail(email);
        if (user.isPresent()) {
            return Response.ok(user.get()).build();
        }
        return Response.status(Response.Status.NOT_FOUND).build();
    }
    
    @GET
    @Path("/search")
    public Response searchUsers(
            @QueryParam("pattern") String pattern,
            @QueryParam("limit") @DefaultValue("10") int limit) {
        List<UserEntity> users = userService.findByNameLike(pattern, limit);
        return Response.ok(users).build();
    }
    
    @GET
    @Path("/status/{status}")
    public Response getUsersByStatus(@PathParam("status") String status) {
        List<UserEntity> users = userService.findByStatus(status);
        return Response.ok(users).build();
    }
    
    @GET
    public Response listUsers(@QueryParam("shardKey") String shardKey) {
        if (shardKey == null) {
            return Response.status(Response.Status.BAD_REQUEST)
                    .entity("shardKey parameter is required").build();
        }
        List<UserEntity> users = userService.listAllUsers(shardKey);
        return Response.ok(users).build();
    }
}

