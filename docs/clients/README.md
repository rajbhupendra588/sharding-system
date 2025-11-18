# Client Libraries & Integration

This section contains documentation for client libraries and integration guides.

## 🚀 NEW: Low-Code Java Client (90-99% Less Code!)

**Want to write 90-99% less code?** Check out our new low-code approach:

- **[Low-Code Guide](./LOW_CODE_GUIDE.md)** - Complete guide to low-code development
- **[Low-Code Summary](./LOW_CODE_SUMMARY.md)** - Implementation overview

**Example:**
```java
// Just define an interface - that's it!
@ShardingRepository(entity = UserEntity.class)
public interface UserRepository extends CrudRepository<UserEntity, String> {
    Optional<UserEntity> findByEmail(String email);
}

// Use it - ONE LINE per operation!
userRepository.findById(id);
userRepository.save(user);
```

## Documentation

- **[How To Guide](./HOW_TO_GUIDE.md)** ⭐⭐⭐ - **COMPLETE!** Step-by-step guide for every use case
- **[Low-Code Guide](./LOW_CODE_GUIDE.md)** ⭐ - **NEW!** 90-99% less code
- **[Naming Conventions](./NAMING_CONVENTIONS.md)** - Use any naming style you want
- **[Java Client Quick Start](./JAVA_QUICKSTART.md)** - Traditional Java client guide
- **[Java Client Reference](./java.md)** - Java client documentation
- **[Quarkus Integration](./QUARKUS_INTEGRATION.md)** - Quarkus integration guide
- **[Quarkus Example](./quarkus-example.md)** - Quarkus example project

## Quick Links

- **Need step-by-step instructions?** Start with [How To Guide](./HOW_TO_GUIDE.md) ⭐⭐⭐
- **New to low-code?** Start with [Low-Code Guide](./LOW_CODE_GUIDE.md) ⭐
- **Wondering about naming?** See [Naming Conventions](./NAMING_CONVENTIONS.md)
- Using Java? Start with [Java Client Quick Start](./JAVA_QUICKSTART.md)
- Integrating with Quarkus? See [Quarkus Integration](./QUARKUS_INTEGRATION.md)
- Need examples? Check [Quarkus Example](./quarkus-example.md)

