plugins {
    kotlin("jvm") version "2.0.20"
    kotlin("plugin.serialization") version "2.0.20"
    id("org.graalvm.buildtools.native") version "0.10.2"
    application
}

group = "com.booktracker"
version = "1.0.0"

repositories {
    mavenCentral()
}

dependencies {
    // http4k
    implementation(platform("org.http4k:http4k-bom:5.10.4.0"))
    implementation("org.http4k:http4k-core")
    implementation("org.http4k:http4k-server-netty")
    implementation("org.http4k:http4k-format-kotlinx-serialization")
    implementation("org.http4k:http4k-client-okhttp")
    
    // Database - Exposed DSL only (no DAO layer to avoid reflection)
    implementation("org.jetbrains.exposed:exposed-core:0.44.1")
    implementation("org.jetbrains.exposed:exposed-jdbc:0.44.1")
    implementation("org.xerial:sqlite-jdbc:3.44.1.0")
    implementation("com.zaxxer:HikariCP:5.1.0")
    
    // Turso libSQL driver for remote SQLite
    implementation("com.dbeaver.jdbc:com.dbeaver.jdbc.driver.libsql:1.0.2")
    
    // Database migrations
    implementation("org.flywaydb:flyway-core:10.4.1")
    
    // Auth & Security
    implementation("org.mindrot:jbcrypt:0.4")
    implementation("com.auth0:java-jwt:4.4.0")
    
    // Serialization
    implementation("org.jetbrains.kotlinx:kotlinx-serialization-json:1.6.0")
    implementation("org.jetbrains.kotlinx:kotlinx-datetime:0.4.1")
    
    // Coroutines for async database operations
    implementation("org.jetbrains.kotlinx:kotlinx-coroutines-core:1.7.3")
    
    // Logging
    implementation("ch.qos.logback:logback-classic:1.4.12")
    
    // Testing
    testImplementation("org.http4k:http4k-testing-hamkrest")
    testImplementation("org.http4k:http4k-testing-kotest")
    testImplementation("org.jetbrains.kotlin:kotlin-test")
    testImplementation("org.junit.jupiter:junit-jupiter:5.10.0")
    testImplementation("io.kotest:kotest-runner-junit5:5.8.0")
    testImplementation("io.kotest:kotest-assertions-core:5.8.0")
    testImplementation("io.mockk:mockk:1.13.8")
    testImplementation("com.h2database:h2:2.2.224")
}

application {
    mainClass.set("com.booktracker.MainKt")
}

kotlin {
    jvmToolchain(21)
}


graalvmNative {
    binaries {
        named("main") {
            imageName.set("book-tracker")
            mainClass.set("com.booktracker.MainKt")
            debug.set(false)
            verbose.set(false) // Reduce log output to save memory
            
            // Memory optimization for Render.com (512MB limit)
            buildArgs.add("-J-Xmx400m") // Very low heap for Render.com
            buildArgs.add("-J-XX:MaxDirectMemorySize=64m") // Minimal direct memory
            
            // Unlock experimental options first
            buildArgs.add("-H:+UnlockExperimentalVMOptions")
            
            // Disable expensive optimizations for faster, lower-memory build
            buildArgs.add("-O0") // Disable optimizations
            buildArgs.add("--no-fallback") // Fail fast if native image can't be built
            buildArgs.add("-H:-UseServiceLoaderFeature") // Disable service loader scanning
            
            buildArgs.add("--initialize-at-build-time=kotlin")
            buildArgs.add("--initialize-at-build-time=kotlinx.serialization")
            buildArgs.add("--initialize-at-build-time=kotlinx.datetime")
            buildArgs.add("--initialize-at-build-time=ch.qos.logback")
            buildArgs.add("-H:+InstallExitHandlers")
            buildArgs.add("-H:+ReportUnsupportedElementsAtRuntime")
            buildArgs.add("-H:+ReportExceptionStackTraces")
            
            // Resource configuration
            buildArgs.add("-H:IncludeResources=.*\\.properties")
            buildArgs.add("-H:IncludeResources=.*\\.conf")
            
            runtimeArgs.add("-Xmx64m")
        }
    }
}

tasks.test {
    useJUnitPlatform()
}

