package com.booktracker.model

import org.jetbrains.exposed.dao.id.LongIdTable

// Users table definition  
object Users : LongIdTable("users") {
    val email = varchar("email", 255).uniqueIndex()
    val passwordHash = varchar("password_hash", 255)
    val firstName = varchar("first_name", 100)
    val lastName = varchar("last_name", 100)
    val isAdmin = bool("is_admin").default(false)
    val createdAt = varchar("created_at", 50)
    val updatedAt = varchar("updated_at", 50)
}

// Children table definition
object Children : LongIdTable("children") {
    val name = varchar("name", 100)
    val age = integer("age")
    val ownerId = long("owner_id")
    val createdAt = varchar("created_at", 50)
    val updatedAt = varchar("updated_at", 50)
}

// Books table definition
object Books : LongIdTable("books") {
    val title = varchar("title", 255)
    val author = varchar("author", 255)
    val dateRead = varchar("date_read", 20)
    val childId = long("child_id")
    val createdAt = varchar("created_at", 50)
    val updatedAt = varchar("updated_at", 50)
}

// Permissions table definition
object Permissions : LongIdTable("permissions") {
    val userId = long("user_id")
    val childId = long("child_id")
    val permissionType = varchar("permission_type", 10)
    val createdAt = varchar("created_at", 50)
}

// Data classes for type safety
data class User(
    val id: Long = 0,
    val email: String,
    val passwordHash: String,
    val firstName: String,
    val lastName: String,
    val isAdmin: Boolean = false,
    val createdAt: String,
    val updatedAt: String
)

data class Child(
    val id: Long = 0,
    val name: String,
    val age: Int,
    val ownerId: Long,
    val createdAt: String,
    val updatedAt: String
)

data class Book(
    val id: Long = 0,
    val title: String,
    val author: String,
    val dateRead: String,
    val childId: Long,
    val createdAt: String,
    val updatedAt: String
)

data class Permission(
    val id: Long = 0,
    val userId: Long,
    val childId: Long,
    val permissionType: String,
    val createdAt: String
)

enum class PermissionType {
    VIEW, EDIT
}