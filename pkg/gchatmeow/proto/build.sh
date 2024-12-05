#!/bin/sh
protoc --go_out=. --go_opt=paths=source_relative --go_opt=embed_raw=true googlechat.proto
goimports -w googlechat.pb.go
