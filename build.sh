docker build \
	-t nanoserverless/nanoserverless \
	-f Dockerfile.build \
	--build-arg http_proxy=$http_proxy \
	--build-arg https_proxy=$https_proxy \
	.
id=$(docker create nanoserverless/nanoserverless)
docker cp $id:/go/src/nanoserverless nanoserverless
docker rm -f $id
docker build \
  -t nanoserverless/nanoserverless:light \
  --build-arg http_proxy=$http_proxy \
  --build-arg https_proxy=$https_proxy \
  .
docker tag nanoserverless/nanoserverless nanoserverless/nanoserverless:dev
