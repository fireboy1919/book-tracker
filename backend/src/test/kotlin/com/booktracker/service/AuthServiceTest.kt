package com.booktracker.service

import com.booktracker.config.TestDatabaseConfig
import com.booktracker.model.LoginRequest
import io.kotest.core.spec.style.DescribeSpec
import io.kotest.matchers.shouldBe
import io.kotest.matchers.shouldNotBe
import io.kotest.matchers.string.shouldStartWith

class AuthServiceTest : DescribeSpec({
    
    beforeEach {
        TestDatabaseConfig.setupTestDatabase()
    }
    
    afterEach {
        TestDatabaseConfig.cleanupTestDatabase()
    }
    
    describe("AuthService") {
        
        describe("hashPassword") {
            it("should hash passwords consistently") {
                val password = "testPassword123"
                val hash1 = AuthService.hashPassword(password)
                val hash2 = AuthService.hashPassword(password)
                
                hash1 shouldNotBe password
                hash2 shouldNotBe password
                hash1 shouldNotBe hash2 // BCrypt uses salt, so hashes should be different
            }
            
            it("should create BCrypt compatible hashes") {
                val password = "testPassword123"
                val hash = AuthService.hashPassword(password)
                
                hash shouldStartWith "\$2a\$"
            }
        }
        
        describe("login") {
            it("should return null for non-existent user") {
                val loginRequest = LoginRequest("nonexistent@example.com", "password")
                val result = AuthService.login(loginRequest)
                
                result shouldBe null
            }
            
            it("should return null for wrong password") {
                // First create a user
                val passwordHash = AuthService.hashPassword("correctPassword")
                UserService.createUser(
                    com.booktracker.model.CreateUserRequest(
                        email = "test@example.com",
                        password = "correctPassword",
                        firstName = "Test",
                        lastName = "User"
                    ),
                    passwordHash
                )
                
                val loginRequest = LoginRequest("test@example.com", "wrongPassword")
                val result = AuthService.login(loginRequest)
                
                result shouldBe null
            }
            
            it("should return login response for correct credentials") {
                // First create a user
                val password = "correctPassword"
                val passwordHash = AuthService.hashPassword(password)
                UserService.createUser(
                    com.booktracker.model.CreateUserRequest(
                        email = "test@example.com",
                        password = password,
                        firstName = "Test",
                        lastName = "User"
                    ),
                    passwordHash
                )
                
                val loginRequest = LoginRequest("test@example.com", password)
                val result = AuthService.login(loginRequest)
                
                result shouldNotBe null
                result!!.token shouldNotBe ""
                result.user.email shouldBe "test@example.com"
                result.user.firstName shouldBe "Test"
                result.user.lastName shouldBe "User"
            }
        }
        
        describe("verifyToken") {
            it("should return null for invalid token") {
                val result = AuthService.verifyToken("invalid.token.here")
                result shouldBe null
            }
            
            it("should return user ID for valid token") {
                // Create user and login to get valid token
                val password = "testPassword"
                val passwordHash = AuthService.hashPassword(password)
                val user = UserService.createUser(
                    com.booktracker.model.CreateUserRequest(
                        email = "test@example.com",
                        password = password,
                        firstName = "Test",
                        lastName = "User"
                    ),
                    passwordHash
                )
                
                val loginRequest = LoginRequest("test@example.com", password)
                val loginResponse = AuthService.login(loginRequest)!!
                
                val userId = AuthService.verifyToken(loginResponse.token)
                userId shouldBe user.id
            }
        }
    }
})