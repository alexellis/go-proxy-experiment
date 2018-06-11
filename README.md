# go-proxy-experiment
Experiments with different methods to proxy requests in Go to upstream URL

# Experiments

## Deploy a sample

Deploy a Golang HTTP function, which has the highest throughput. One exists already (`alexellis2/gofast30-5-18`):

```
faas deploy --image=alexellis2/gofast30-5-18 --name=gofast30-5-18
```

> Also pass ` --gateway=$GW_URL`

## Deploy the proxy

Deploy the go-proxy-experiment on Kubernetes.

```
kubectl apply -f ./yaml/dep.yaml
```

Then pick a NodePort or LB:

```
kubectl apply -f ./yaml/nodeport-svc.yaml

# or

kubectl apply -f ./yaml/lb-svc.yaml
```

## Collect results

Now try each variation of the proxy via a different route.

The $URL variable is the NodePort:31115 or LB:8080

Use `http.Post`

```
~/go/bin/hey -d "" -m POST -c 200 -n 20000 http://$URL/http-post?fn=gofast30-5-18.openfaas-fn.svc.cluster.local.:8080 -d test -m POST
```

Reuse the same client instance with `http.NewRequest`/`http.Do`:

```
~/go/bin/hey -d "" -m POST -c 200 -n 20000 http://$URL/client-post?fn=gofast30-5-18.openfaas-fn.svc.cluster.local.:8080 -d test -m POST
```

Use the reverse proxy implementation in OpenFaaS based upon `http.NewRequest`/`http.Do`:

```
~/go/bin/hey -d "" -m POST -c 200 -n 20000 http://$URL/faas-post?fn=gofast30-5-18.openfaas-fn.svc.cluster.local.:8080 -d test -m POST
```
