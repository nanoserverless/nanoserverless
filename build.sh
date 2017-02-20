curl -L "https://github.com/msoap/shell2http/releases/download/1.8/shell2http-1.8.linux.amd64.tar.gz" | tar xzf - shell2http
chmod 700 shell2http
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
rm -f shell2http
