main:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w -extldflags "-static"'

dist: main
	upx --brute cdn77-refresh

