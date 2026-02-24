# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates libc6-compat wget

WORKDIR /root/ems/backend/internal/scheduling
COPY scheduling .

VOLUME ["/root/ems/backend/internal/common/config"]

EXPOSE 8083 9003

CMD ["./scheduling"]
