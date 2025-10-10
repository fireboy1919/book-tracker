package com.booktracker.config

import com.booktracker.model.*
import org.jetbrains.exposed.sql.Database
import org.jetbrains.exposed.sql.SchemaUtils
import org.jetbrains.exposed.sql.deleteAll
import org.jetbrains.exposed.sql.transactions.transaction

object TestDatabaseConfig {
    private var database: Database? = null
    
    fun setupTestDatabase() {
        if (database == null) {
            database = Database.connect(
                url = "jdbc:h2:mem:test;DB_CLOSE_DELAY=-1;MODE=PostgreSQL",
                driver = "org.h2.Driver"
            )
            
            transaction {
                SchemaUtils.create(
                    Users,
                    Children,
                    Books,
                    Permissions
                )
            }
        } else {
            cleanupTestDatabase()
        }
    }
    
    fun cleanupTestDatabase() {
        transaction {
            // Delete in order to respect foreign key constraints
            Books.deleteAll()
            Permissions.deleteAll()
            Children.deleteAll()
            Users.deleteAll()
        }
    }
}