#!/bin/bash

set -e

etcdctl rm --recursive /akins.org/onedari || true


for i in {0..100}; do
    curl --fail -H "Content-Type: application/json" -svo /dev/null -X PUT --data-binary '{"port": 9999, "up": true, "ip":"10.10.10.10", "labels":{"datacenter": "east", "app":"my_app"}}}' http://127.0.0.1:63412/v0/instances/$i
    curl  --fail -svo /dev/null http://127.0.0.1:63412/v0/instances/$i .
done


curl  --fail -H "Content-Type: application/json" -svo /dev/null -X PUT --data-binary '{"labels":{"app":"my_app"}, "query": {"app":"my_app", "datacenter": "east"}}}' http://127.0.0.1:63412/v0/services/my_app


time curl  --fail -sv  http://127.0.0.1:63412/v0/services/my_app | jq . | wc -l

