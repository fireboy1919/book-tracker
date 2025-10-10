package com.booktracker

import org.flywaydb.core.Flyway
import kotlin.system.exitProcess

/**
 * Simple migration runner that doesn't depend on SQLDelight
 */
fun main() {
    try {
        val jdbcUrl = "jdbc:sqlite:./booktracker.db"
        
        val flyway = Flyway.configure()
            .dataSource(jdbcUrl, null, null)
            .locations("classpath:db/migration")
            .load()
        
        println("Running database migrations...")
        val result = flyway.migrate()
        println("Migrations completed: ${result.migrationsExecuted} migrations executed")
        println("Database created successfully at ./booktracker.db")
        
    } catch (e: Exception) {
        println("Migration failed: ${e.message}")
        e.printStackTrace()
        exitProcess(1)
    }
}