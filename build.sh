docker run -it \
	--rm -v $PWD:/go/src \
	-w /go/src \
	-e http_proxy=$http_proxy \
	-e https_proxy=$http_proxy \
	-e "GO_PATH=/go/src" \
	-e "CGO_ENABLED=0" golang:1.7 \
	sh -c "go get github.com/gorilla/mux github.com/docker/docker/client github.com/docker/docker/api/types && go build -a --installsuffix cgo --ldflags=-s -o nanoserverless" \
&& docker build -t nanoserverless .
