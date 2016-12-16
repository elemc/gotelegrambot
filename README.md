[![Build Status](https://travis-ci.org/elemc/gotelegrambot.svg?branch=master)](https://travis-ci.org/elemc/gotelegrambot)
[![Go Report Card](https://goreportcard.com/badge/github.com/elemc/gotelegrambot)](https://goreportcard.com/report/github.com/elemc/gotelegrambot)
[![GoDoc](https://godoc.org/github.com/elemc/gotelegrambot?status.svg)](https://godoc.org/github.com/elemc/gotelegrambot)

gotelegrambot
=============

This is simple bot for Telegram writen on Go (golang) and use Couchbase for data store.
Example server http://logs.elemc.name bot store logs from chats

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
