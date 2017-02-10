# nanoserverless
<a href="https://hub.docker.com/r/nanoserverless/nanoserverless" target="blank"><img src="https://upload.wikimedia.org/wikipedia/commons/7/79/Docker_(container_engine)_logo.png" height="20"/></a>
[![Build Status](https://travis-ci.org/nanoserverless/nanoserverless.svg?branch=master)](https://travis-ci.org/nanoserverless/nanoserverless)
[![Go Report Card](https://goreportcard.com/badge/nanoserverless/nanoserverless)](http://goreportcard.com/report/nanoserverless/nanoserverless)
[![](https://images.microbadger.com/badges/image/nanoserverless/nanoserverless.svg)](https://microbadger.com/images/nanoserverless/nanoserverless)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/nanoserverless/nanoserverless/blob/master/LICENSE.md)
<a href="https://travis-ci.org/nanoserverless/nanoserverless" target="blank"><img src="https://travis-ci.org/nanoserverless/nanoserverless.svg?branch=master" height="20"/></a>

## Example
You can test that on http://play-with-docker.com

### Swarm init if needed
```
docker swarm init
```

### Up service
```
docker network create -d overlay nanoserverless
docker service create \
  --name nanoserverless \
  --mount type=bind,source=/var/run/docker.sock,destination=/var/run/docker.sock \
  --network nanoserverless \
  --publish 1664:80 \
  nanoserverless/nanoserverless:master
```

### Create pi function in node7 (time to build FROM node:7 image)
```
time curl 'http://<ip_manager>:<port>/node7/pi/create?url=https://raw.githubusercontent.com/nanoserverless/nanoserverless/master/examples/pi/pi.js'
real    0m6.701s
```

### Exec that function (in serverless mode)
```
time curl 'http://<ip_manager>:<port>/node7/pi/exec'
3.1415926445762157
real    0m0.891s
```

### Up a service for that function
```
time curl 'http://<ip_manager>:<port>/node7/pi/up'
Service id  pylnihmv8w0ymuf3ovniuzuvb created
real    0m0.061s
```

### Exec that function (in service mode now)
```
time curl 'http://<ip_manager>:<port>:10080/node7/pi/exec'
3.1415926445762157
real    0m0.440s
```

### Down service
```
time curl 'http://<ip_manager>:<port>/node7/pi/down'
Service nanoserverless-node7-pi removed
real    0m0.015s
```

## Next
### Multiple nodes
If you have a swarm cluster, you need to place the REGISTRY environment variable  
Creating function will push image to that registry  
So other nodes can pull if needed  
