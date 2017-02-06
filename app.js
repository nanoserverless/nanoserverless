var express = require('express');
var morgan = require('morgan');
var querystring = require('querystring');
var dockerode = require('dockerode');
var tar = require('tar-stream');
var dockermodem = require('docker-modem');
var zlib = require('zlib');
var concat = require('concat-stream');

var tagprefix = 'nanoserverless-';

// Docker modem
var modem = new dockermodem();

// Dockerode
var docker = new dockerode();

// Express
var app = express();
app.use(morgan('combined'));

app.get('/function', function (req, res) {
  res.send('NanoServerLess');
})

app.get('/function/create/:base', function (req, res) {
  var base = req.params.base;

  // Test tar from pipe
  var code = 'phpinfo();';
  var pack = tar.pack();
  pack.entry({ name: 'Dockerfile' }, 'FROM	php:7\nCOPY app.php /\nENTRYPOINT ["php", "app.php"]');
  pack.entry({ name: 'app.php' }, '<?php\n' + code + '\n?>');
  pack.finalize();
  pack.pipe(zlib.createGzip()).pipe(concat(function (file) {
    var opts = {t: tagprefix + "montag", nocache: "true"};
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
      modem.followProgress(data, onFinished, onProgress);
      function onFinished(err, output) {
        res.end();
      }
      function onProgress(err, output) {
        res.write(err.stream);
      }
    });
  }));
})

app.get('/function/exec/:tag', function (req, res) {
  var tag = req.params.tag;
  docker.run(tagprefix + tag, [querystring.stringify(req.query)], res, function (err, data, container) { });
});

app.listen(3000, function () {
  console.log('NanoServerLess started on port 3000');
});
