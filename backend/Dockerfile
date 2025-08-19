# --- Builder stage ---
FROM golang:1.24-alpine AS builder

WORKDIR /src
RUN apk add --no-cache ca-certificates git tzdata

COPY go.mod go.sum ./
RUN go mod download
COPY . .
ENV CGO_ENABLED=0

RUN mkdir -p /out

RUN go build -trimpath -ldflags "-s -w -buildid=" -o /out/server ./cmd/server \
 && go build -trimpath -ldflags "-s -w -buildid=" -o /out/import-offers ./cmd/import-offers


# --- Production stage ---
FROM gcr.io/distroless/base-debian12:nonroot AS prod

WORKDIR /app
COPY --from=builder /out/server /app/server
COPY --from=builder /out/import-offers /app/import-offers

ARG DATABASE_URL=""
ENV PORT=8090 \
	TZ=Etc/UTC \
	DATABASE_URL=${DATABASE_URL}

EXPOSE 8090
USER nonroot:nonroot

ENTRYPOINT ["/app/server"]

