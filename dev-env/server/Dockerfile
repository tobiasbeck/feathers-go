# base image
FROM golang:1.16.0-alpine

WORKDIR /usr/src/
 
RUN go get github.com/githubnemo/CompileDaemon
ENV APP_ENV=devcontainer

# start app
ENTRYPOINT CompileDaemon --directory="./" --build="go build cmd/main.go" --command=./main