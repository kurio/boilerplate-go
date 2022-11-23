FROM golang:1.19-alpine3.16 as builder

ENV GOPRIVATE github.com/KurioApp
WORKDIR /app

RUN apk update && \
    apk upgrade && \
    apk add --update --no-cache alpine-sdk && \
    rm -rf /var/cache/apk/*

ARG GITHUB_OAUTH
RUN git config --global url."https://${GITHUB_OAUTH}:x-oauth-basic@github.com/".insteadOf "https://github.com/"

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o goboilerplate github.com/kurio/boilerplate-go/cmd/goboilerplate

FROM gcr.io/distroless/static
WORKDIR /app
EXPOSE 7723

COPY --from=builder /app/goboilerplate .

CMD ["./goboilerplate", "http"]
