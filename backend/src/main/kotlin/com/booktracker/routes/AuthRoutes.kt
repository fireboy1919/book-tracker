package com.booktracker.routes

import com.booktracker.model.*
import com.booktracker.service.AuthService
import com.booktracker.service.UserService
import org.http4k.core.*
import org.http4k.format.KotlinxSerialization.auto
import org.http4k.routing.bind
import org.http4k.routing.routes

fun authRoutes() = routes(
    "/register" bind Method.POST to ::registerUser,
    "/login" bind Method.POST to ::loginUser
)

private val createUserLens = Body.auto<CreateUserRequest>().toLens()
private val loginLens = Body.auto<LoginRequest>().toLens()
private val loginResponseLens = Body.auto<LoginResponse>().toLens()
private val userResponseLens = Body.auto<UserResponse>().toLens()
private val errorResponseLens = Body.auto<ErrorResponse>().toLens()

private fun registerUser(request: Request): Response {
    return try {
        val createUserRequest = createUserLens(request)
        
        // Check if user already exists
        val existingUser = UserService.getUserByEmail(createUserRequest.email)
        if (existingUser != null) {
            return Response(Status.BAD_REQUEST)
                .with(errorResponseLens of ErrorResponse("User with this email already exists"))
        }
        
        // Hash password and create user
        val passwordHash = AuthService.hashPassword(createUserRequest.password)
        val user = UserService.createUser(createUserRequest, passwordHash)
        
        val userResponse = UserResponse(
            id = user.id,
            email = user.email,
            firstName = user.firstName,
            lastName = user.lastName,
            isAdmin = user.isAdmin,
            createdAt = user.createdAt
        )
        
        Response(Status.CREATED).with(userResponseLens of userResponse)
        
    } catch (e: Exception) {
        Response(Status.INTERNAL_SERVER_ERROR)
            .with(errorResponseLens of ErrorResponse("Registration failed: ${e.message}"))
    }
}

private fun loginUser(request: Request): Response {
    return try {
        val loginRequest = loginLens(request)
        
        val loginResponse = AuthService.login(loginRequest)
        if (loginResponse != null) {
            Response(Status.OK).with(loginResponseLens of loginResponse)
        } else {
            Response(Status.UNAUTHORIZED)
                .with(errorResponseLens of ErrorResponse("Invalid credentials"))
        }
        
    } catch (e: Exception) {
        Response(Status.INTERNAL_SERVER_ERROR)
            .with(errorResponseLens of ErrorResponse("Login failed: ${e.message}"))
    }
}