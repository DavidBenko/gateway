#!/usr/bin/env bash

./wrap.rb routes.js | curl -X PUT -d @- localhost:5000/admin/routes 
./wrap.rb endpoint_hi.js id=1 name=hi | curl -X PUT -d @- localhost:5000/admin/proxy_endpoints/1
