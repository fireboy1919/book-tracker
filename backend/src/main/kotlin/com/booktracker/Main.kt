package com.booktracker

import com.booktracker.config.DatabaseConfig
import com.booktracker.routes.authRoutes
import com.booktracker.routes.userRoutes
import com.booktracker.routes.childRoutes
import com.booktracker.routes.bookRoutes
import com.booktracker.routes.testRoutes
import org.http4k.core.*
import org.http4k.filter.CorsPolicy
import org.http4k.filter.ServerFilters
import org.http4k.routing.bind
import org.http4k.routing.routes
import org.http4k.server.Netty
import org.http4k.server.asServer

fun main() {
    // Initialize database
    DatabaseConfig.initDatabase()
    
    val port = System.getenv("PORT")?.toInt() ?: 8080
    
    val app = ServerFilters.Cors(
        CorsPolicy.UnsafeGlobalPermissive
    ).then(
        routes(
            "/api/auth" bind authRoutes(),
            "/api/users" bind userRoutes(),
            "/api/children" bind childRoutes(),
            "/api/books" bind bookRoutes(),
            "/api/test" bind testRoutes(),
            "/health" bind Method.GET to { Response(Status.OK).body("OK") }
        )
    )
    
    println("Starting server on port $port")
    app.asServer(Netty(port)).start()
}