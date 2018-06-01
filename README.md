# go-proxy-experiment
Experiments with different methods to proxy requests in Go to upstream URL

# Experiments

Deploy a Golang HTTP function: 

`alexellis2/gofast30-5-18`


Now try each variation of the proxy via a different route:

Use `http.Post`

```
export $GW = 127.0.0.1:31112
~/go/bin/hey -d "" -m POST -c 200 -n 20000 http://$IP/http-post?fn=gofast30-5-18.openfaas-fn.svc.cluster.local.:8080 -d test -m POST
```

Reuse the same client instance with `http.NewRequest`/`http.Do`:

```
export $GW = 127.0.0.1:31112
~/go/bin/hey -d "" -m POST -c 200 -n 20000 http://$IP/client-post?fn=gofast30-5-18.openfaas-fn.svc.cluster.local.:8080 -d test -m POST
```

Use the reverse proxy implementation in OpenFaaS based upon `http.NewRequest`/`http.Do`:

```
export $GW = 127.0.0.1:31112
~/go/bin/hey -d "" -m POST -c 200 -n 20000 http://$IP/faas-post?fn=gofast30-5-18.openfaas-fn.svc.cluster.local.:8080 -d test -m POST
```

