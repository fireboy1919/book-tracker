package com.booktracker.routes

import com.booktracker.config.TestDatabaseConfig
import com.booktracker.model.CreateUserRequest
import com.booktracker.model.LoginRequest
import io.kotest.core.spec.style.DescribeSpec
import io.kotest.matchers.shouldBe
import org.http4k.core.Method
import org.http4k.core.Request
import org.http4k.core.Status
import org.http4k.core.with
import org.http4k.format.KotlinxSerialization.auto

class AuthRoutesTest : DescribeSpec({
    
    val createUserLens = org.http4k.core.Body.auto<CreateUserRequest>().toLens()
    val loginLens = org.http4k.core.Body.auto<LoginRequest>().toLens()
    
    beforeEach {
        TestDatabaseConfig.setupTestDatabase()
    }
    
    afterEach {
        TestDatabaseConfig.cleanupTestDatabase()
    }
    
    describe("Auth Routes") {
        val routes = authRoutes()
        
        describe("POST /register") {
            it("should register a new user successfully") {
                val createUserRequest = CreateUserRequest(
                    email = "test@example.com",
                    password = "password123",
                    firstName = "Test",
                    lastName = "User"
                )
                
                val request = Request(Method.POST, "/register")
                    .with(createUserLens of createUserRequest)
                
                val response = routes(request)
                
                response.status shouldBe Status.CREATED
            }
            
            it("should reject registration with duplicate email") {
                val createUserRequest = CreateUserRequest(
                    email = "test@example.com",
                    password = "password123",
                    firstName = "Test",
                    lastName = "User"
                )
                
                val request = Request(Method.POST, "/register")
                    .with(createUserLens of createUserRequest)
                
                // First registration should succeed
                routes(request).status shouldBe Status.CREATED
                
                // Second registration with same email should fail
                val response2 = routes(request)
                response2.status shouldBe Status.BAD_REQUEST
            }
        }
        
        describe("POST /login") {
            it("should login with correct credentials") {
                // First register a user
                val createUserRequest = CreateUserRequest(
                    email = "test@example.com",
                    password = "password123",
                    firstName = "Test",
                    lastName = "User"
                )
                
                routes(Request(Method.POST, "/register")
                    .with(createUserLens of createUserRequest))
                
                // Then try to login
                val loginRequest = LoginRequest(
                    email = "test@example.com",
                    password = "password123"
                )
                
                val request = Request(Method.POST, "/login")
                    .with(loginLens of loginRequest)
                
                val response = routes(request)
                
                response.status shouldBe Status.OK
            }
            
            it("should reject login with wrong password") {
                // First register a user
                val createUserRequest = CreateUserRequest(
                    email = "test@example.com",
                    password = "password123",
                    firstName = "Test",
                    lastName = "User"
                )
                
                routes(Request(Method.POST, "/register")
                    .with(createUserLens of createUserRequest))
                
                // Try to login with wrong password
                val loginRequest = LoginRequest(
                    email = "test@example.com",
                    password = "wrongpassword"
                )
                
                val request = Request(Method.POST, "/login")
                    .with(loginLens of loginRequest)
                
                val response = routes(request)
                
                response.status shouldBe Status.UNAUTHORIZED
            }
            
            it("should reject login for non-existent user") {
                val loginRequest = LoginRequest(
                    email = "nonexistent@example.com",
                    password = "password123"
                )
                
                val request = Request(Method.POST, "/login")
                    .with(loginLens of loginRequest)
                
                val response = routes(request)
                
                response.status shouldBe Status.UNAUTHORIZED
            }
        }
    }
})