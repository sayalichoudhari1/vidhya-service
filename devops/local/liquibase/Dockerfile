FROM liquibase/liquibase:4.33.0

USER 0:0

ENV DB_HOST=postgres
ENV DB_PORT=5432
ENV DB_NAME=vidhyadb

RUN apt-get update && apt-get install -y gettext

COPY ./docker-entrypoint-custom.sh /liquibase/docker-entrypoint-custom.sh
COPY ./init-scripts/* /dbtool/scripts/

RUN chmod +x /liquibase/docker-entrypoint-custom.sh /dbtool/scripts/*

ENTRYPOINT [ "/liquibase/docker-entrypoint-custom.sh" ]
CMD [ "update" ]
