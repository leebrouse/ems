# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates libc6-compat wget

WORKDIR /root/ems/backend/internal/warehouse
COPY warehouse .

VOLUME ["/root/ems/backend/internal/common/config"]

EXPOSE 8082 9002

CMD ["./warehouse"]
