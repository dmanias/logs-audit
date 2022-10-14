# logs-audit

This is a logs audit API. Events from logs are aggregated and the user can run queries on them.
Events are indexed by their invariant parts and the variant parts are stored all together under the name data.
The endpoints are protected with bearer token authentication.

For install run the setup.sh
set up git
```shell
sudo apt install git
```
clone https://github.com/dmanias/logs-audit
```shell
clone git@github.com:dmanias/logs-audit.git
```
run set up script
```shell
sudo chmod u+x setup.shsudo 
./setup.sh
```

searchDBHandler
storeEventsHandler
authenticationHandler
registrationsHandler

In the test/test_endpoints.sh are the curl calls for the API.
For the search operation the results are exported in the benchmarks.txt file and the metrics are presented in the console.

Unit tests are in the main and auth folders.