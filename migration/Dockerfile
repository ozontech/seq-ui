FROM gomicro/goose

WORKDIR /migration

COPY *.sql ./
COPY entrypoint.sh ./

RUN chmod +x ./entrypoint.sh

ENTRYPOINT ["./entrypoint.sh"]