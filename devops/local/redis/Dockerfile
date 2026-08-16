FROM redis:8.8.0-alpine

ENV REDIS_HOST=redis
ENV REDIS_PORT=6379

RUN apk upgrade --no-cache
COPY ./redis.conf /usr/local/etc/redis/redis.conf

CMD ["redis-server", "/usr/local/etc/redis/redis.conf"]