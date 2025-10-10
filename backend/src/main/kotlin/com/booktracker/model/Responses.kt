package com.booktracker.model

import kotlinx.serialization.Serializable

@Serializable
data class UserResponse(
    val id: Long,
    val email: String,
    val firstName: String,
    val lastName: String,
    val isAdmin: Boolean,
    val createdAt: String
)

@Serializable
data class LoginResponse(
    val token: String,
    val user: UserResponse
)

@Serializable
data class ChildResponse(
    val id: Long,
    val name: String,
    val age: Int,
    val ownerId: Long,
    val createdAt: String
)

@Serializable
data class BookResponse(
    val id: Long,
    val title: String,
    val author: String,
    val dateRead: String,
    val childId: Long,
    val createdAt: String
)

@Serializable
data class PermissionResponse(
    val id: Long,
    val userId: Long,
    val childId: Long,
    val permissionType: String,
    val createdAt: String
)

@Serializable
data class ErrorResponse(
    val message: String,
    val code: String? = null
)