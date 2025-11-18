# Naming Conventions - You Choose!

## Important: No Naming Restrictions

The Sharding System Java Client **does NOT enforce any naming conventions**. You can use **any names** you want for:

- Entity classes
- Repository interfaces  
- Service classes
- Field names
- Method names (except CRUD method names from `CrudRepository`)

## Examples with Different Naming Styles

### Example 1: Domain-Driven Design Style

```java
// Entity - can be named anything
@Entity(table = "customer_accounts")
public class CustomerAccount {
    @ShardKey
    private String accountNumber;
    private String customerName;
    // ...
}

// Repository - can be named anything
@ShardingRepository(entity = CustomerAccount.class)
public interface CustomerAccountRepository extends CrudRepository<CustomerAccount, String> {
    Optional<CustomerAccount> findByAccountNumber(String accountNumber);
}

// Service - can be named anything
@ApplicationScoped
public class AccountManagementService {
    @Inject
    CustomerAccountRepository accountRepository;
    
    public Optional<CustomerAccount> getAccount(String accountNumber) {
        return accountRepository.findById(accountNumber);
    }
}
```

### Example 2: Simple/Short Names

```java
// Entity
@Entity(table = "orders")
public class Order {
    @ShardKey
    private String orderId;
    // ...
}

// Repository
@ShardingRepository(entity = Order.class)
public interface OrderRepo extends CrudRepository<Order, String> {
    List<Order> findByCustomerId(String customerId);
}

// Service
@ApplicationScoped
public class OrderService {
    @Inject
    OrderRepo orderRepo;
    // ...
}
```

### Example 3: Hungarian/Abbreviated Style

```java
// Entity
@Entity(table = "product_inventory")
public class ProdInv {
    @ShardKey
    private String prodId;
    // ...
}

// Repository
@ShardingRepository(entity = ProdInv.class)
public interface ProdInvRepo extends CrudRepository<ProdInv, String> {
    // ...
}
```

### Example 4: Full Descriptive Names

```java
// Entity
@Entity(table = "financial_transaction_records")
public class FinancialTransactionRecord {
    @ShardKey
    private String transactionIdentifier;
    // ...
}

// Repository
@ShardingRepository(entity = FinancialTransactionRecord.class)
public interface FinancialTransactionRecordRepository 
    extends CrudRepository<FinancialTransactionRecord, String> {
    // ...
}
```

### Example 5: Your Company's Naming Convention

```java
// Entity
@Entity(table = "tbl_users")
public class TblUser {
    @ShardKey
    private String userId;
    // ...
}

// Repository
@ShardingRepository(entity = TblUser.class)
public interface TblUserRepository extends CrudRepository<TblUser, String> {
    // ...
}
```

## What Matters: Annotations, Not Names

The framework only cares about:

1. **Annotations** - `@Entity`, `@ShardKey`, `@ShardingRepository`, etc.
2. **Table names** - Specified in `@Entity(table = "...")` or auto-generated
3. **Column names** - Specified in `@Column(name = "...")` or auto-generated from field names
4. **CRUD method names** - Must match `CrudRepository` interface (save, findById, etc.)

## Field Name → Column Name Mapping

Field names are automatically converted to snake_case for column names:

```java
@Entity
public class MyEntity {
    @ShardKey
    private String userId;        // → column: "user_id"
    
    private String emailAddress;  // → column: "email_address"
    
    private String firstName;     // → column: "first_name"
    
    @Column(name = "custom_col")  // → column: "custom_col" (overrides auto-naming)
    private String customField;
}
```

## Custom Table Names

You can override table names regardless of class name:

```java
@Entity(table = "my_custom_table_name")
public class AnythingYouWant {
    // Class name doesn't matter, table name is "my_custom_table_name"
}
```

## Custom Column Names

You can override column names regardless of field name:

```java
@Entity
public class MyClass {
    @ShardKey
    @Column(name = "id")
    private String myIdField;  // Maps to column "id", not "my_id_field"
    
    @Column(name = "full_name")
    private String name;  // Maps to column "full_name", not "name"
}
```

## Repository Interface Names

Repository interfaces can be named **anything**:

```java
// All of these work:
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {}

@ShardingRepository(entity = User.class)
public interface UserRepo extends CrudRepository<User, String> {}

@ShardingRepository(entity = User.class)
public interface UserDAO extends CrudRepository<User, String> {}

@ShardingRepository(entity = User.class)
public interface IUserService extends CrudRepository<User, String> {}

@ShardingRepository(entity = User.class)
public interface WhateverYouWant extends CrudRepository<User, String> {}
```

## Custom Query Method Names

Custom query methods can be named **anything** (except CRUD methods):

```java
@ShardingRepository(entity = User.class)
public interface UserRepository extends CrudRepository<User, String> {
    // All of these work:
    Optional<User> findByEmail(String email);
    Optional<User> getByEmail(String email);
    Optional<User> lookupByEmail(String email);
    Optional<User> fetchUserByEmail(String email);
    Optional<User> findUserWithEmail(String email);
    
    // But you need @Query annotation for non-standard patterns
    @Query("SELECT * FROM users WHERE email = $1")
    Optional<User> customMethodName(String email);
}
```

## Summary

✅ **Use any class names** - `User`, `CustomerAccount`, `ProdInv`, `TblUser`, etc.  
✅ **Use any repository names** - `UserRepository`, `UserRepo`, `UserDAO`, etc.  
✅ **Use any field names** - `id`, `userId`, `accountNumber`, etc.  
✅ **Use any method names** - `findByEmail`, `getByEmail`, `lookupByEmail`, etc.  
✅ **Override table/column names** - Via `@Entity(table = "...")` and `@Column(name = "...")`  

❌ **Only requirement:** Use correct annotations (`@Entity`, `@ShardKey`, `@ShardingRepository`)

The framework is **completely flexible** with naming - use whatever conventions your team prefers!

