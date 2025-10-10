package com.booktracker.config

import com.booktracker.model.*
import com.zaxxer.hikari.HikariConfig
import com.zaxxer.hikari.HikariDataSource
import org.jetbrains.exposed.sql.Database
import org.jetbrains.exposed.sql.SchemaUtils
import org.jetbrains.exposed.sql.transactions.transaction
import org.flywaydb.core.Flyway

object DatabaseConfig {
    
    fun initDatabase(): Database {
        val config = HikariConfig()
        
        // Get database URL from environment
        val databaseUrl = System.getenv("DATABASE_URL") ?: "file:./booktracker.db"
        
        when {
            // Turso libSQL URL format: libsql://database-name.turso.io?authToken=token
            databaseUrl.startsWith("libsql://") -> {
                config.driverClassName = "org.libsql.jdbc.LibsqlDriver"
                config.jdbcUrl = databaseUrl.replace("libsql://", "jdbc:libsql://")
                config.maximumPoolSize = 5 // Turso can handle more connections
            }
            // Local SQLite file
            databaseUrl.startsWith("file:") -> {
                config.driverClassName = "org.sqlite.JDBC"
                config.jdbcUrl = "jdbc:sqlite:${databaseUrl.substring(5)}"
                config.maximumPoolSize = 3
            }
            // Direct JDBC URL (already formatted)
            databaseUrl.startsWith("jdbc:") -> {
                when {
                    databaseUrl.contains("sqlite") -> {
                        config.driverClassName = "org.sqlite.JDBC"
                        config.maximumPoolSize = 3
                    }
                    databaseUrl.contains("libsql") -> {
                        config.driverClassName = "org.libsql.jdbc.LibsqlDriver"
                        config.maximumPoolSize = 5
                    }
                    else -> throw IllegalArgumentException("Unsupported database URL: $databaseUrl")
                }
                config.jdbcUrl = databaseUrl
            }
            else -> {
                // Default to local SQLite
                config.driverClassName = "org.sqlite.JDBC"
                config.jdbcUrl = "jdbc:sqlite:./booktracker.db"
                config.maximumPoolSize = 3
            }
        }
        
        // Add connection properties for better performance
        config.addDataSourceProperty("cachePrepStmts", "true")
        config.addDataSourceProperty("prepStmtCacheSize", "250")
        config.addDataSourceProperty("prepStmtCacheSqlLimit", "2048")
        
        val dataSource = HikariDataSource(config)
        
        // Run Flyway migrations
        runMigrations(config.jdbcUrl)
        
        return Database.connect(dataSource)
    }
    
    private fun runMigrations(jdbcUrl: String) {
        println("Running database migrations for: ${jdbcUrl.substringBefore("?")}")
        
        val flyway = Flyway.configure()
            .dataSource(jdbcUrl, null, null)
            .locations("classpath:db/migration")
            .load()
        
        try {
            flyway.migrate()
            println("Database migrations completed successfully")
        } catch (e: Exception) {
            println("Migration failed: ${e.message}")
            throw e
        }
    }
}