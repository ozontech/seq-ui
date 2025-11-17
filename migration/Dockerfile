FROM gomicro/goose

COPY migration/*.sql /migrations
COPY migration/entrypoint.sh /migrations

RUN chmod +x /migrations/entrypoint.sh

ENTRYPOINT ["/migrations/entrypoint.sh"]