[![Build Status](https://travis-ci.org/elemc/gotelegrambot.svg?branch=master)](https://travis-ci.org/elemc/gotelegrambot)

gotelegrambot
=============

This is simple bot for Telegram writen on Go (golang) and use Couchbase for data store.

Compile
-------

### Requires
- golang >= 1.5.1 (http://www.golang.org)
- git
- installed Couchbase cluster (http://couchbase.com)
- github.com/couchbase/gocb
- gopkg.in/telegram-bot-api.v4
- github.com/gin-gonic/gin

### Download
- $ github.com/couchbase/gocb gopkg.in/telegram-bot-api.v4 github.com/gin-gonic/gin
- $ go get github.com/elemc/gotelegrambot

### Build
$ go build github.com/elemc/gotelegrambot
