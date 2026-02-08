FROM debian:bookworm-slim
USER root

WORKDIR /app
RUN mkdir -p /app/data/db

RUN apt-get update && \
    apt-get install -y --no-install-recommends ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/lib/apt/lists/*

COPY . .
EXPOSE 80

CMD ["./Goptime"]