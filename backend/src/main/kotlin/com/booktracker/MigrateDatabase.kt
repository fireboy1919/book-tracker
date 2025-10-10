package com.booktracker

import org.flywaydb.core.Flyway

/**
 * Utility to run Flyway migrations manually
 * This creates the database schema that SQLDelight will then read
 */
fun main() {
    val jdbcUrl = "jdbc:sqlite:./booktracker.db"
    
    val flyway = Flyway.configure()
        .dataSource(jdbcUrl, null, null)
        .locations("classpath:db/migration")
        .load()
    
    println("Running database migrations...")
    val result = flyway.migrate()
    println("Migrations completed: ${result.migrationsExecuted} migrations executed")
}