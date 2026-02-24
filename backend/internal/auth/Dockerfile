FROM alpine:latest
RUN apk --no-cache add ca-certificates libc6-compat wget

WORKDIR /root/ems/backend/internal/auth
COPY auth .

VOLUME ["/root/ems/backend/internal/common/config"]

EXPOSE 8080 9000

CMD ["./auth"]