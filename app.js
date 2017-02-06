var express = require('express');
var morgan = require('morgan');
var querystring = require('querystring');
var dockeroder = require('dockerode');

var app = express();

app.use(morgan('combined'));

app.get('/function', function (req, res) {
  res.send('NanoServerLess')
})

app.listen(3000, function () {
  console.log('NanoServerLess started on port 3000');
});
