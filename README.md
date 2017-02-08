# nanoserverless
Docker Hub : https://hub.docker.com/r/nanoserverless/nanoserverless/

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
curl -X POST -H 'Content-Type: text/plain' \
  'http://localhost:1664/create/node7/showparams' \
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

### Exec php7 func new image
```
docker run --rm -e "key1=val1" -e "key2=val2" nanoserverless-php7-showparams | python -m json.tool
```
Result :
```
{
    "GPG_KEYS": "A917B1ECDA84AEC2B568FED6F50ABC807BD5DCD0 528995BFEDFBA7191D46839EF9BA0ADA31CBD89E",
    "HOME": "/root",
    "HOSTNAME": "395d6cfbc24d",
    "MAVAR": "titi",
    "PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
    "PHPIZE_DEPS": "autoconf \t\tfile \t\tg++ \t\tgcc \t\tlibc-dev \t\tmake \t\tpkg-config \t\tre2c",
    "PHP_ASC_URL": "https://secure.php.net/get/php-7.1.1.tar.xz.asc/from/this/mirror",
    "PHP_CFLAGS": "-fstack-protector-strong -fpic -fpie -O2",
    "PHP_CPPFLAGS": "-fstack-protector-strong -fpic -fpie -O2",
    "PHP_INI_DIR": "/usr/local/etc/php",
    "PHP_LDFLAGS": "-Wl,-O1 -Wl,--hash-style=both -pie",
    "PHP_MD5": "65eef256f6e7104a05361939f5e23ada",
    "PHP_SHA256": "b3565b0c1441064eba204821608df1ec7367abff881286898d900c2c2a5ffe70",
    "PHP_URL": "https://secure.php.net/get/php-7.1.1.tar.xz/from/this/mirror",
    "PHP_VERSION": "7.1.1",
    "TERM": "xterm"
}
```

### Exec node7 func new image
```
docker run --rm -e "key1=val1" -e "key2=val2" nanoserverless-node7-showparams | python -m json.tool
```
Result :
```
{
    "HOME": "/root",
    "HOSTNAME": "5861331356c9",
    "NODE_VERSION": "7.5.0",
    "NPM_CONFIG_LOGLEVEL": "info",
    "PATH": "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
    "key1": "val1",
    "key2": "val2"
}
```

## TODO
### Exec php7 func
```
curl 'http://localhost:1664/exec/php7/showparams?p1=parm1&p2=parm2' | python -m json.tool
```

### Exec node7 func
```
curl 'http://localhost:1664/exec/node7/showparams?p1=parm1&p2=parm2' | python -m json.tool
```
