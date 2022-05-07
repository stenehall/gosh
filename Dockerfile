FROM golang:1.17-alpine AS builder

ARG BUILD_VERSION
ARG BUILD_DATE

WORKDIR /app

RUN apk add --update-cache alpine-sdk upx

COPY go.mod go.sum Makefile ./
COPY internal ./internal
COPY cmd ./cmd

RUN go mod download
RUN make fetch-fontawesome
RUN CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=${BUILD_VERSION}" -o build/gosh ./cmd/gosh
RUN upx --best --lzma -o /app/gosh /app/build/gosh

FROM gcr.io/distroless/static AS final

ARG BUILD_VERSION
ARG BUILD_DATE

LABEL maintainer="Johan Stenehall"
LABEL org.label-schema.build-date=$BUILD_DATE
LABEL org.label-schema.application=gosh
LABEL org.label-schema.version=$BUILD_VERSION

ENV APP_ENV production

USER nonroot:nonroot

WORKDIR /
COPY --from=builder --chown=nonroot:nonroot /app/gosh /gosh
COPY --from=builder --chown=nonroot:nonroot /app/assets /assets
COPY --chown=nonroot:nonroot web /web

HEALTHCHECK --interval=60s --timeout=5s --start-period=2s --retries=3 CMD ["/gosh", "-health"]

EXPOSE 8080

ENTRYPOINT [ "/gosh" ]
