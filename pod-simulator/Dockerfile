FROM golang:alpine as build
WORKDIR /build
COPY . .
RUN go build -o pod-simulator main.go

FROM alpine
COPY --from=build /build/pod-simulator /usr/local/bin/pod-simulator
ENTRYPOINT ["pod-simulator"]