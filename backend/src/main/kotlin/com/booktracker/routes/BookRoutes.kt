package com.booktracker.routes

import org.http4k.core.*
import org.http4k.routing.bind
import org.http4k.routing.routes

fun bookRoutes() = routes(
    "/" bind Method.GET to { Response(Status.OK).body("Book routes - TODO") }
)