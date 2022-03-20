# syntax=docker/dockerfile:1

###################### Build ######################
FROM golang:1.17.8-alpine AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY *.go ./

RUN go build -o /msgbroker

###################### Deploy ######################
FROM alpine:3.15.1

# Copy our msgbroker executable from the deploy stage
WORKDIR /
COPY --from=build /msgbroker /msgbroker

EXPOSE 3000

# Finally run the executable
CMD ["/msgbroker"]
