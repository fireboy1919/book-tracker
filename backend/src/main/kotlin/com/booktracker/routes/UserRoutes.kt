package com.booktracker.routes

import org.http4k.core.*
import org.http4k.routing.bind
import org.http4k.routing.routes

fun userRoutes() = routes(
    "/" bind Method.GET to { Response(Status.OK).body("User routes - TODO") }
)