#! /usr/bin/sh
curl -X PUT "http://localhost:8180/put?key=a&value=b"
curl "http://localhost:8180/get?key=a"
