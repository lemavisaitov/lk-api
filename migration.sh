#!/bin/bash
source .env

export MIGRATION_DSN="host=postgres port=5432 user=${DB_USER} password=${DB_PASSWORD} sslmode=disable"

goose -dir "./migrations" postgres "${MIGRATION_DSN}" up -v