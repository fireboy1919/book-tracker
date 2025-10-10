package com.booktracker.routes

import com.booktracker.model.*
import org.http4k.core.*
import org.http4k.routing.bind
import org.http4k.routing.routes
import org.jetbrains.exposed.sql.deleteAll
import org.jetbrains.exposed.sql.transactions.transaction

fun testRoutes() = routes(
    "/reset-db" bind Method.DELETE to ::resetDatabase
)

private fun resetDatabase(request: Request): Response {
    return try {
        transaction {
            // Delete all data in order to respect foreign key constraints
            Books.deleteAll()
            Permissions.deleteAll()
            Children.deleteAll()
            Users.deleteAll()
        }
        Response(Status.OK).body("Database reset successfully")
    } catch (e: Exception) {
        Response(Status.INTERNAL_SERVER_ERROR).body("Failed to reset database: ${e.message}")
    }
}