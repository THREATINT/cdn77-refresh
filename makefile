cdn77-refresh: cdn77-refresh.go
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-s -w -extldflags "-static"'
	upx --brute cdn77-refresh
