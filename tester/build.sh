#!/usr/bin/env sh

cp -r ../rtc/web ./web

docker build -t tester .

rm -rf web
