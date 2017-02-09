# nanoserverless
Docker Hub : https://hub.docker.com/r/nanoserverless/nanoserverless/

## Up in seconds
```
docker run -d \
  -p 1664:3000 \
  -v /var/run/docker.sock:/var/run/docker.sock \
  --name nanoserverless nanoserverless/nanoserverless:master-light
```

## Up in swarm mode
docker network create -d overlay nanoserverless
docker service create \
  --name nanoserverless \
  --mount type=bind,source=/var/run/docker.sock,destination=/var/run/docker.sock \
  --network nanoserverless \
  --publish 1664:80 \
  nanoserverless/nanoserverless:master-light

### Create php7 func
```
curl -X POST \
  'http://<ip_docker>:1664/php7/showparams/create' \
  -d 'print(json_encode($_ENV));'
```
Result :
```
Image  nanoserverless-php7-showparams created !

Dockerfile:
FROM php:7
COPY app /
ENTRYPOINT ["php", "app"]

Code:
<?php
print(json_encode($_ENV));
?>

Log:
{"stream":"sha256:b7cf52ddccaeb003270e4b513037d64847baeb3061063bb367545945a2d99ecf\n"}
```

### Create node7 func
```
curl -X POST \
  'http://<ip_docker>:1664/node7/showparams/create' \
  -d 'console.log(JSON.stringify(process.env));'
```
Result :
```
Image  nanoserverless-node7-showparams created !

Dockerfile:
FROM node:7
COPY app /
ENTRYPOINT ["node", "app"]

Code:
console.log(JSON.stringify(process.env));


Log:
{"stream":"sha256:e4b329b11f9d355273f643e1b275b613a075ddc2d38d830254cb48f6a861404c\n"}
```

### Exec php7 func
```
curl 'http://<ip_docker>:1664/php7/showparams/exec?p1=parm1&p2=parm2' | python -m json.tool
```

### Exec node7 func
```
curl 'http://<ip_docker>:1664/node7/showparams/exec?p1=parm1&p2=parm2' | python -m json.tool
```
