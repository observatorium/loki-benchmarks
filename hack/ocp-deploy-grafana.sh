#!/bin/bash

set -eou pipefail

NAMESPACE="${1:-observatorium-logs-test}"
TENANT_ID="${2:-observatorium}"
LOKI_PREFIX="${3:-observatorium-loki}"

# Deploy grafana and expose as route
oc --ignore-not-found=true -n "$NAMESPACE" delete deployment grafana
oc create deployment -n "$NAMESPACE" grafana --image=docker.io/grafana/grafana
oc --ignore-not-found=true -n "$NAMESPACE" delete service grafana
oc expose -n "$NAMESPACE" deployment grafana --port=3000
oc --ignore-not-found=true -n "$NAMESPACE" delete route grafana
oc expose -n "$NAMESPACE" service grafana
grafana_url=$(oc get -n "$NAMESPACE" route grafana --no-headers | awk {'print $2'})

# Expose query_frontend and distributor http services as routes
oc --ignore-not-found=true -n "$NAMESPACE" delete route "$LOKI_PREFIX"-distributor-http
oc expose -n "$NAMESPACE" service "$LOKI_PREFIX"-distributor-http
loki_distributor_url=$(oc -n "$NAMESPACE" get route "$LOKI_PREFIX"-distributor-http --no-headers | awk {'print $2'})

oc --ignore-not-found=true -n "$NAMESPACE" delete route "$LOKI_PREFIX"-query-frontend-http
oc expose -n "$NAMESPACE" service "$LOKI_PREFIX"-query-frontend-http
loki_query_frontend_url=$(oc -n "$NAMESPACE" get route "$LOKI_PREFIX"-query-frontend-http --no-headers | awk {'print $2'})

# Add loki datasource to grafana (using loki_query_frontend_url)
curl -i -XPOST --silent -u admin:admin -H "Content-Type: application/json" -H "Accept: application/json" "http://${grafana_url}/api/datasources" -d '
{
  "name": "loki",
  "type": "loki",
  "basicAuth": false,
  "access": "proxy",
  "url":"http://'"${loki_query_frontend_url}":80'",
  "jsonData": {
    "httpHeaderName1": "X-Scope-OrgID"
  },
  "secureJsonData":  {
    "httpHeaderValue1": "'"${TENANT_ID}"'"
  }
}'

# User instructions
echo -e "\n\n"
echo "To access grafana use:"
echo "--=-=-==--=-=-=-=-==--"
echo "http://${grafana_url}"
echo "user: admin password: admin"
echo -e "\n"
echo -e "Note: in grafana explore tab change datasource to Loki and run query such as {client=\"promtail\"} \n"
