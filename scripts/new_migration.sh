#!/usr/bin/env bash

migrate create -ext sql -dir migrations -seq "$1"
