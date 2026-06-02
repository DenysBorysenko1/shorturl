#!/usr/bin/env bash

POSTGRESQL_URL="${POSTGRESQL_URL:-postgres://postgres:postgres@localhost:5432/yndx_go?sslmode=disable}"

migrate -database ${POSTGRESQL_URL} -path migrations up
