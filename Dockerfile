FROM golang:1.27-bookworm AS build
ARG CMD=widedoor
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/app ./cmd/${CMD}

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/app /app
COPY --from=build /src/data /data
USER 65532:65532
ENTRYPOINT ["/app"]
