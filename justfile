alias s := start
alias l := lint
alias t := test

default:
    just --list

# start the client and server
start:
    hivemind

# format & lint the server code
lint:
    cd server && go fmt ./... && golangci-lint run ./... && nilaway ./...

# format and check the server code
test:
    cd server && go test ./...
