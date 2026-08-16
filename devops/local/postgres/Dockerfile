FROM postgres:18.4-alpine

ENV DB_HOST=postgres
ENV DB_PORT=5432
ENV DB_NAME=vidhyadb

ENV DB_MASTER_USER=appuser
ENV DB_MASTER_PASS=appuser
ENV DB_USER=appuser
ENV DB_PASS=appuser

RUN apk upgrade --no-cache
RUN apk add gettext

COPY ./init.sh /docker-entrypoint-initdb.d/init.sh
