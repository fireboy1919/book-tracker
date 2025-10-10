package com.booktracker.routes

import org.http4k.core.*
import org.http4k.routing.bind
import org.http4k.routing.routes

fun childRoutes() = routes(
    "/" bind Method.GET to { Response(Status.OK).body("Child routes - TODO") }
)