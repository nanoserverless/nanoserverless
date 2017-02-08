# nanoserverless

## Up in seconds
```
docker run -d \
  -p 1664:3000 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  --name nanoserverless nanoserverless/nanoserverless
```

### Create php7 func
```
curl -X POST -H 'Content-Type: text/plain' \
  'http://localhost:1664/create/php7/showparams' \
  -d 'print(json_encode($_ENV));'
```

### Create node7 func
```
curl -X POST -H 'Content-Type: text/plain' \
  'http://localhost:1664/create/node7/showparams' \
  -d 'console.log(JSON.stringify(process.env));'
```

### Exec php7 func
```
curl 'http://localhost:1664/exec/php7/showparams?p1=parm1&p2=parm2'
```

### Exec node7 func
```
curl 'http://localhost:1664/exec/node7/showparams?p1=parm1&p2=parm2'
```
