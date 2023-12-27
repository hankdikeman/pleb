# TODO would be nice to make some of this Dockerfile stuff reusable. I could:
#   1. have a base image builder, with gRPC/proto stuff preloaded
#   2. have a single Dockerfile template for each microservice
#   3. modularize some of these commands, maybe a grpc-setup.sh?

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
RUN go build -C ./main -v -o /usr/local/bin/horrea .

###
# PRODUCTION STAGE
###
FROM debian:bullseye AS prod

# Copy binary from build stage to production image
COPY --from=build /usr/local/bin/horrea /usr/local/bin/horrea

# Expose working port and set entrypoint command
EXPOSE 55412
ENTRYPOINT ["horrea"]
