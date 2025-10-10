# Multi-stage build for GraalVM native image
FROM ghcr.io/graalvm/graalvm-ce:21 AS builder

# Install native-image
RUN gu install native-image

WORKDIR /app
COPY . .

# Build the application
RUN ./gradlew clean build

# Build native image
RUN ./gradlew nativeCompile

# Runtime stage - use distroless for minimal attack surface
FROM gcr.io/distroless/base-debian12:nonroot

WORKDIR /app

# Copy the native binary
COPY --from=builder /app/build/native/nativeCompile/book-tracker /app/book-tracker

# Note: distroless doesn't have mkdir, so data directory will be created by mount
# The directory will be mounted at runtime by Render

# Expose port (Render will override with PORT env var)
EXPOSE 10000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/app/book-tracker", "--health-check"] || exit 1

# Run the native binary
ENTRYPOINT ["/app/book-tracker"]