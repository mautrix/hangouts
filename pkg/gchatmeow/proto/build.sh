#!/bin/sh
protoc --go_out=gchatproto --go_opt=paths=source_relative --go_opt=embed_raw=true googlechat.proto
protoc --go_out=gchatprotoweb --go_opt=paths=source_relative --go_opt=embed_raw=true googlechatweb.proto
goimports -w */*.pb.go
