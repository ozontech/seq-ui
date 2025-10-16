#!/bin/bash

DBSTRING="host=$DBHOST port=$DBPORT user=$DBUSER password=$DBPASSWORD dbname=$DBNAME sslmode=$DBSSL"

goose postgres "$DBSTRING" $MIGRATIONCOMMAND