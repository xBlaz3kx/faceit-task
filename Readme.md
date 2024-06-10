# ESL Faceit Group Technical task

## Description

The project is a simple user management service, that allows to create, read, update, delete users as well as watching
for user changes (streaming). The service is implemented using Go and gRPC API. The data is stored in MongoDB. The
service also contains an HTTP health check endpoint, that can be used to check the service status.

The solution is fully containerized using Docker and Docker compose.

## Deployment

The application can be deployed using Docker or running the binary directly.

### Using Docker Compose (recommended)

Deploy the solution using Docker compose:

```bash
docker-compose up --build -d
```

The deployment will start both the service and MongoDB containers. The service's gRPC API will be available on
port 8080 by default, but can be configured using the config file or environment variables.

### Running the binary

To run the service directly, you need to have a MongoDB instance running. The service can be started using the following
command:

```bash
  USER_DATABASE_URI=mongodb://mongo:mongo@localhost:27017/faceit?authSource=admin go run cmd/user-service/main.go 
```

## Configuration

The application can be configured using a configuration file or through environment variables. The following options can
be configured:

- `server` - the server address and port
- `database` - the MongoDB connection string

### Configuration file

The application configuration is stored in the `config.yaml` file. The configuration file contains the following
fields:

```yaml
# The gRPC server address and port
server: 0.0.0.0:8080
# The MongoDB connection string
database: mongodb://mongo:mongo@localhost:27017
```

### Environment variables

All environment variables are prefixed with `USER` prefix. The underscore (`_`) is used as a delimited to separate any
nested fields. For example, the `server` configuration can be set using the `USER_SERVER` environment variable.

### Flags

To see the available flags, run the service with the `--help` flag. The output will be similar to the following:

```text
User service

Usage:
user [flags]

Flags:
-c, --config string   Path to the configuration file
-h, --help            help for user
-v, --version         version for user
```

## Project Structure

Using the Clean Architecture and Domain Driven Design principles, the project is structured in the following way:

```
.
├── build/
│ └── Dockerfile # Dockerfile for building the service with mod and build cache
├── cmd/
│ └── user-service/
│ └── main.go # Includes a Cobra root command for running the service
└── internal/
    ├── api/
    │ ├── grpc/
    │ │ └── ... # GRPC server abstraction and GRPC handlers
    │ └── http/
    │   └── ... # HTTP server with healthcheck
    ├── domain/
    │   └── services/
    │       └── users.go # Service interface and implementation
    ├── repositories/
    │   ├── mongo/
    │   │ ├── main.go
    │   │ └── users-repository.go # Mongo repository implementation
    │   └── users.go # User repository model and interface definition
    └── pkg/
        ├── configuration # Application configuration file
        └── proto/
            └── v1/
                ├── user.proto # gRPC proto definition and compiled files (service + messages)
                └── ... (compiled proto files)
```

## Notes

- Passwords are hashed using bcrypt and stored in the database.
- As no specific validation requirements were provided, I concluded that minimal validation should be present - when
  creating a user, a valid email and password with a minimum length of 8 characters are required.
- Getting a user or listing users won't return the password hash in the response object (for security reasons).
- Simplified change streams - the service currently emits changes to multiple GRPC clients using an internal
  notification/messaging system. Changes are emitted in the service level, after the database operation is successful.
- Currently, all changes are emitted to all clients. This could be improved by adding a filter to the change stream.
- The health checks are implemented using the HTTP API. The healthcheck endpoint is available at `/healthz`. This
  could've been implemented using gRPC as well.
- TLS certificate handling is not implemented, but should be added for production use.
- Both GRPC and HTTP API are using logging and recovery middleware/interceptors. In production, full observability would
  be nice as well.