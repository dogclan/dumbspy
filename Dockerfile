FROM golang:1.22.8-alpine AS build

ARG build_commit_sha="-"
ARG build_version="-"
ARG build_time="-"

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN mkdir -p /app/src  \
    && mkdir -p /app/bin

WORKDIR /app/src

COPY go.mod go.sum ./
RUN go mod download &&  \
    go mod verify

COPY . ./

RUN go build -v \
    -o /app/bin/dumbspy \
    -ldflags="-X 'main.buildTime=$build_time' -X 'main.buildCommit=$build_commit_sha' -X 'main.buildVersion=$build_version'" \
    /app/src/cmd/dumbspy

FROM gcr.io/distroless/base-debian11

WORKDIR /

COPY --from=build /app/bin/dumbspy /dumbspy

EXPOSE 29900

USER nonroot:nonroot

CMD [ "/dumbspy" ]
