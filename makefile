amd64: cdn77-refresh.go go.mod
	CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -a -ldflags '-s -w -extldflags "-static"'

upx:
	upx --brute cdn77-refresh

deps:
	go get -u all

clean:
	rm -f cdn77-refresh cdn77-refresh.upx

