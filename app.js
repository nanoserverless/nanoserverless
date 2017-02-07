var express = require('express');
var morgan = require('morgan');
var querystring = require('querystring');
var dockerode = require('dockerode');
var tar = require('tar-stream');
var dockermodem = require('docker-modem');
var zlib = require('zlib');
var concat = require('concat-stream');
var fs = require('fs');

// Globals
var shell2http_url = 'https://github.com/msoap/shell2http/releases/download/1.7/shell2http-1.7.linux.amd64.zip';
var tagprefix = 'nanoserverless-';

// Docker modem
var modem = new dockermodem();

// Dockerode
var docker = new dockerode();

// Express
var app = express();
app.use(morgan('combined'));
app.use(function(req, res, next) {
  var contentType = req.headers['content-type'] || ''
  , mime = contentType.split(';')[0];
  if (mime != 'text/plain') {
    return next();
  }
  var data = '';
  req.setEncoding('utf8');
  req.on('data', function(chunk) {
    data += chunk;
  });
  req.on('end', function() {
    req.rawBody = data;
    next();
  });
});


app.get('/', function (req, res) {
  res.send('NanoServerLess');
});

var dockerfiles = {
  "php7": {
    "from": "php:7",
    "file": "app.php",
    "cmd": "php",
    "precode": '<?php\nparse_str($argv[1], $params);',
    "postcode": '?>'
  },
  "node7": {
    "from": "node:7",
    "file": "app.js",
    "cmd": "node",
    "precode": 'var querystring = require(\'querystring\');\nvar params = querystring.parse(process.argv[2]);'
  }
};

app.post('/create/:base/:name', function (req, res) {
  var base = req.params.base;
  var name = req.params.name;

  // Init pack
  var pack = tar.pack();

  // Dockerfile
  var dockerfile =
    'FROM ' + dockerfiles[base].from +
    '\nCOPY shell2http /' +
    '\nCOPY ' + dockerfiles[base].file + ' /' +
    '\nENTRYPOINT ["' + dockerfiles[base].cmd + '", "' + dockerfiles[base].file + '"]';
  pack.entry({ name: 'Dockerfile' }, dockerfile);
  
  // App Code
  var code = '';
  if (req.rawBody) code = req.rawBody;
  pack.entry({ name: dockerfiles[base].file}, dockerfiles[base].precode + '\n' + code + '\n' + dockerfiles[base].postcode);

  // shell2http
  /*var entry = pack.entry({ name: 'shell2http' }, function(err) {
    pack.finalize();
  });
  fs.createReadStream('shell2http').on('data', function (chunk) {
    entry.write(chunk);
  });*/

  var data = '';
  var stream = fs.createReadStream('shell2http');

  stream.on('data', function(chunk) {
    data += chunk;
  });


  stream.on('end', function()  {
      pack.entry({ name: 'shell2http' }, data);
      pack.finalize();
      
      // Tag
      var tag = tagprefix + base + '-' + name;

      pack.pipe(zlib.createGzip()).pipe(concat(function (file) {
        var opts = {t: tag, nocache: "true"};
        var optsf = {
          path: '/build?',
          method: 'POST',
          file: file,
          options: opts,
          isStream: true,
          //openStdin: true,
          statusCodes: {
            200: true,
            500: 'server error'
          }
        };

        res.writeHead(200, {
          'Content-Type': 'text/plain',
          'Connection': 'Transfer-Encoding',
          'Transfer-Encoding': 'chunked'
        });

        console.log('Creating ' + tag);
        modem.dial(optsf, function(err, data) {
          if (err) console.log(err);
          modem.followProgress(data,
            function(err, output) {
              // Finished
              res.end();
            },
            function(err, output) {
              // Progress
              if (err.stream) res.write(err.stream);
            }
          );
        });
      }));
    });
})

app.get('/exec/:base/:name', function (req, res) {
  var base = req.params.base;
  var name = req.params.name;
  var tag = tagprefix + base + '-' + name;
  console.log('Running ' + tag);
  docker.run(tag, querystring.stringify(req.query), res, {}, function (err, data, container) {
    // Remove container
    docker.getContainer(container.id).remove({}, function(err, data) { });
  });
});

app.listen(3000, function () {
  console.log('NanoServerLess started on port 3000');
});
