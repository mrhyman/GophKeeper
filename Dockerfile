FROM docker-hosted.artifactory.tcsbank.ru/cicd-images/golang-1.26:latest AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG VERSION=dev
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-s -w \
    -X gophkeeper/pkg/buildinfo.Version=${VERSION} \
    -X gophkeeper/pkg/buildinfo.BuildDate=${BUILD_DATE}" \
    -o /bin/server ./cmd/server


FROM docker-hosted.artifactory.tcsbank.ru/cicd-images/golang-1.26:latest AS migrator

ENV GOBIN=/usr/local/bin
RUN go install github.com/pressly/goose/v3/cmd/goose@latest


FROM docker-hosted.artifactory.tcsbank.ru/cicd-images/base-jammy:latest AS final

RUN apt-get update && \
    apt-get install -y --no-install-recommends \
        ca-certificates \
        tzdata && \
    rm -rf /var/lib/apt/lists/*

RUN groupadd -r gophkeeper && useradd -r -g gophkeeper gophkeeper

WORKDIR /app

COPY --from=builder /bin/server /app/server
COPY --from=migrator /usr/local/bin/goose /usr/local/bin/goose
COPY internal/server/migrations/*.sql /app/migrations/
COPY docker-entrypoint.sh /app/docker-entrypoint.sh
RUN chmod +x /app/docker-entrypoint.sh

COPY configs/server.example.yaml /app/configs/server.yaml

RUN chown -R gophkeeper:gophkeeper /app

USER gophkeeper

EXPOSE 8080

ENTRYPOINT ["/app/docker-entrypoint.sh"]
CMD ["/app/server"]
