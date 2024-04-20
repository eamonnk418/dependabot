FROM golang:1.22.2-alpine3.19 AS builder

WORKDIR /app

COPY go.mod *go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/dependabot ./cmd/dependabot/

FROM scratch

COPY --from=builder /bin/dependabot /bin/dependabot

ENTRYPOINT [ "/bin/dependabot" ]