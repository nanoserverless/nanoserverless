var express = require('express');
var morgan = require('morgan');
var querystring = require('querystring');
var dockerode = require('dockerode');
var tar = require('tar-stream');
var dockermodem = require('docker-modem');
var zlib = require('zlib');
var concat = require('concat-stream');

// Globals
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

  var pack = tar.pack();
  var dockerfile =
    'FROM ' + dockerfiles[base].from +
    '\nCOPY ' + dockerfiles[base].file + ' /' +
    '\nENTRYPOINT ["' + dockerfiles[base].cmd + '", "' + dockerfiles[base].file + '"]';
  pack.entry({ name: 'Dockerfile'}, dockerfile);

  // Test tar from pipe
  console.log(req.rawBody);
  var code = '';
  if (req.rawBody) code = req.rawBody;
  pack.entry({ name: dockerfiles[base].file}, dockerfiles[base].precode + '\n' + code + '\n' + dockerfiles[base].postcode);
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

    modem.dial(optsf, function(err, data) {
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
})

app.get('/exec/:base/:name', function (req, res) {
  var base = req.params.base;
  var name = req.params.name;
  var tag = tagprefix + base + '-' + name;
  docker.run(tag, querystring.stringify(req.query), res, function (err, data, container) {
    console.log(err);
  });
});

app.listen(3000, function () {
  console.log('NanoServerLess started on port 3000');
});
