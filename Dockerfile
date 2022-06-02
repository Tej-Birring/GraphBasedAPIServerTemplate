FROM golang:1.18-bullseye

# info
MAINTAINER tejbirring@gmail.com

# create and set working dir
RUN mkdir /build
WORKDIR /build 

# copy go.mod to working dir
COPY go.mod .
COPY go.sum .

# download dependencies
RUN go mod download 

# copy src to working dir
COPY . .

# set env vars
ENV PORT 80

# build
RUN go build -o /go/bin/GraphBasedServer

# run
ENTRYPOINT ["GraphBasedServer"]