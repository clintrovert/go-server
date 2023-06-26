# go-server

Simple go server with Docker build file

Project structure follows [goland-standards](https://github.com/golang-standards/project-layout) recommendations.

### Pre-requisites

Generate modeling structs for gRPC contract

```bash
protoc --go_out=. --go_opt=paths=source_relative \  
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    api/model/*.proto
```

### Local Development

Spin up local server and dependencies - `docker-compose up -d`

Requesting unary User endpoints - 

#### Get User
```bash
grpcurl -H 'authorization: Bearer test' -d '{"id":"<test>"}' -plaintext localhost:9090 playground.UserService.GetUser
```

#### Create User
```bash
grpcurl -H 'authorization: Bearer test' -d '{"name":"test","email":"test@test.com","password":"helloworld"}' -plaintext localhost:9090 playground.UserService.CreateUser
```

#### Update User
```bash
grpcurl -H 'authorization: Bearer test' -d '{"id":"<test>","name":"updatedName","email":"updatedEmail@password.com", "password":"updatedPassword"}' -plaintext localhost:9090 playground.UserService.UpdateUser
```

#### Delete User
```bash
grpcurl -H 'authorization: Bearer test' -d '{"user_id":"<test>"}' -plaintext localhost:9090 playground.UserService.DeleteUser
```
