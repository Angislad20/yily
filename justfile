alias s := start
alias t := test

default:
    just --list

# start the client and server
start:
    hivemind

# format and check the server code
test:
    cd server && go fmt ./... && go test ./... && golangci-lint run ./... && nilaway ./...
