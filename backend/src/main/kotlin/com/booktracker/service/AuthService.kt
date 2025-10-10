package com.booktracker.service

import com.auth0.jwt.JWT
import com.auth0.jwt.algorithms.Algorithm
import com.auth0.jwt.exceptions.JWTVerificationException
import com.booktracker.model.LoginRequest
import com.booktracker.model.LoginResponse
import com.booktracker.model.UserResponse
import org.mindrot.jbcrypt.BCrypt
import java.time.Instant
import java.time.temporal.ChronoUnit

object AuthService {
    private val jwtSecret = System.getenv("JWT_SECRET") ?: "default-secret-change-me"
    private val algorithm = Algorithm.HMAC256(jwtSecret)
    
    fun hashPassword(password: String): String {
        return BCrypt.hashpw(password, BCrypt.gensalt())
    }
    
    fun verifyPassword(password: String, hash: String): Boolean {
        return BCrypt.checkpw(password, hash)
    }
    
    fun login(request: LoginRequest): LoginResponse? {
        val user = UserService.getUserByEmail(request.email) ?: return null
        
        if (!verifyPassword(request.password, user.passwordHash)) {
            return null
        }
        
        val token = generateToken(user.id)
        val userResponse = UserResponse(
            id = user.id,
            email = user.email,
            firstName = user.firstName,
            lastName = user.lastName,
            isAdmin = user.isAdmin,
            createdAt = user.createdAt
        )
        
        return LoginResponse(token = token, user = userResponse)
    }
    
    private fun generateToken(userId: Long): String {
        val expiresAt = Instant.now().plus(24, ChronoUnit.HOURS)
        
        return JWT.create()
            .withSubject(userId.toString())
            .withExpiresAt(expiresAt)
            .sign(algorithm)
    }
    
    fun verifyToken(token: String): Long? {
        return try {
            val verifier = JWT.require(algorithm).build()
            val jwt = verifier.verify(token)
            jwt.subject.toLong()
        } catch (e: JWTVerificationException) {
            null
        }
    }
}