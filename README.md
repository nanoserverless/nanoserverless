# nanoserverless

## Up in 1 minute
```
docker build -t nanoserverless .
docker rm -f nanoserverless
docker run -d -p 1664:3000 -v /var/run/docker.sock:/var/run/docker.sock --name nanoserverless nanoserverless
```

### Create php7 func
```
curl 'http://localhost:1664/create/php7/showparams' -d 'var_dump($params);'
```

### Create node7 func
```
curl 'http://localhost:1664/create/node7/showparams' -d 'console.log(JSON.stringify(params));'
```

### Exec php7 func
```
curl 'http://localhost:1664/exec/php7/showparams?p1=parm1&p2=parm2'
```

### Exec node7 func
```
curl 'http://localhost:1664/exec/node7/showparams?p1=parm1&p2=parm2'
```
