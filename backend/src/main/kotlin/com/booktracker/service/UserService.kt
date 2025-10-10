package com.booktracker.service

import com.booktracker.model.*
import org.jetbrains.exposed.sql.*
import org.jetbrains.exposed.sql.SqlExpressionBuilder.eq
import org.jetbrains.exposed.sql.transactions.transaction
import kotlinx.datetime.Clock
import kotlinx.datetime.TimeZone
import kotlinx.datetime.toLocalDateTime

object UserService {

    fun createUser(request: CreateUserRequest, passwordHash: String): User {
        return transaction {
            val now = Clock.System.now().toLocalDateTime(TimeZone.UTC).toString()
            
            val userId = Users.insertAndGetId {
                it[email] = request.email
                it[Users.passwordHash] = passwordHash
                it[firstName] = request.firstName
                it[lastName] = request.lastName
                it[isAdmin] = false
                it[createdAt] = now
                it[updatedAt] = now
            }

            User(
                id = userId.value,
                email = request.email,
                passwordHash = passwordHash,
                firstName = request.firstName,
                lastName = request.lastName,
                isAdmin = false,
                createdAt = now,
                updatedAt = now
            )
        }
    }

    fun getUserByEmail(email: String): User? {
        return transaction {
            Users.select { Users.email eq email }
                .singleOrNull()
                ?.let { row ->
                    User(
                        id = row[Users.id].value,
                        email = row[Users.email],
                        passwordHash = row[Users.passwordHash],
                        firstName = row[Users.firstName],
                        lastName = row[Users.lastName],
                        isAdmin = row[Users.isAdmin],
                        createdAt = row[Users.createdAt],
                        updatedAt = row[Users.updatedAt]
                    )
                }
        }
    }

    fun getUserById(id: Long): User? {
        return transaction {
            Users.select { Users.id eq id }
                .singleOrNull()
                ?.let { row ->
                    User(
                        id = row[Users.id].value,
                        email = row[Users.email],
                        passwordHash = row[Users.passwordHash],
                        firstName = row[Users.firstName],
                        lastName = row[Users.lastName],
                        isAdmin = row[Users.isAdmin],
                        createdAt = row[Users.createdAt],
                        updatedAt = row[Users.updatedAt]
                    )
                }
        }
    }

    fun getAllUsers(): List<User> {
        return transaction {
            Users.selectAll().map { row ->
                User(
                    id = row[Users.id].value,
                    email = row[Users.email],
                    passwordHash = row[Users.passwordHash],
                    firstName = row[Users.firstName],
                    lastName = row[Users.lastName],
                    isAdmin = row[Users.isAdmin],
                    createdAt = row[Users.createdAt],
                    updatedAt = row[Users.updatedAt]
                )
            }
        }
    }

    fun updateUser(id: Long, request: UpdateUserRequest): User? {
        return transaction {
            val now = Clock.System.now().toLocalDateTime(TimeZone.UTC).toString()
            
            val updated = Users.update({ Users.id eq id }) {
                it[email] = request.email
                it[firstName] = request.firstName
                it[lastName] = request.lastName
                it[isAdmin] = request.isAdmin
                it[updatedAt] = now
            }

            if (updated > 0) {
                getUserById(id)
            } else {
                null
            }
        }
    }

    fun deleteUser(id: Long): Boolean {
        return transaction {
            Users.deleteWhere { Users.id eq id } > 0
        }
    }
}