FROM golang:latest AS build-env

ARG packages=main.go

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./${packages}


FROM alpine:latest

WORKDIR /app

COPY --from=build-env /app/main .

COPY log-output.yaml .

RUN chmod +x ./main

EXPOSE 8000

CMD ["/app/main"]