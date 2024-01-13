###
# BUILD STAGE
###
FROM --platform=linux/x86_64 golang:1.21-bullseye AS build

# Move to working directory /app
WORKDIR /usr/src/app

# NOTE: to avoid the pain of handling gRPC/proto dependencies here, I'm
#   just requiring that the generated code already be present in ./pb

# Copy and build server code
COPY . .
RUN go build -C ./prod/senator/main -v -o /usr/local/bin/senator .

###
# PRODUCTION STAGE
###
FROM debian:bullseye AS prod

# Copy binary from build stage to production image
COPY --from=build /usr/local/bin/senator /usr/local/bin/senator

# Expose working port and set entrypoint command
EXPOSE ${SENATOR_PORT}
ENTRYPOINT ["senator"]
