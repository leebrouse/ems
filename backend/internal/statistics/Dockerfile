# Run stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates libc6-compat wget

WORKDIR /root/ems/backend/internal/statistics
COPY statistics .

VOLUME ["/root/ems/backend/internal/common/config"]

EXPOSE 8084 9004

CMD ["./statistics"]
