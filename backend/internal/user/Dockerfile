FROM alpine:latest
RUN apk --no-cache add ca-certificates libc6-compat wget

WORKDIR /root/ems/backend/internal/user
COPY user .

VOLUME ["/root/ems/backend/internal/common/config"]

EXPOSE 8081 9001

CMD ["./user"]