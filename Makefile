.PHONY: arm
arm: armv6 armv7

.PHONY: armv6
armv6:
	GOOS=linux GOARCH=arm GOARM=6 go build -o dist/switchbot-meter.linux-armv6l

.PHONY: armv7
armv7:
	GOOS=linux GOARCH=arm GOARM=7 go build -o dist/switchbot-meter.linux-armv7