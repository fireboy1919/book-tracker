package com.booktracker.model

import kotlinx.serialization.Serializable

@Serializable
data class CreateUserRequest(
    val email: String,
    val password: String,
    val firstName: String,
    val lastName: String
)

@Serializable
data class UpdateUserRequest(
    val email: String,
    val firstName: String,
    val lastName: String,
    val isAdmin: Boolean
)

@Serializable
data class LoginRequest(
    val email: String,
    val password: String
)

@Serializable
data class CreateChildRequest(
    val name: String,
    val age: Int
)

@Serializable
data class UpdateChildRequest(
    val name: String,
    val age: Int
)

@Serializable
data class CreateBookRequest(
    val title: String,
    val author: String,
    val dateRead: String,
    val childId: Long
)

@Serializable
data class UpdateBookRequest(
    val title: String,
    val author: String,
    val dateRead: String
)

@Serializable
data class CreatePermissionRequest(
    val userId: Long,
    val childId: Long,
    val permissionType: String
)