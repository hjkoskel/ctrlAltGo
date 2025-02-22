GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o init
echo init | cpio -o --format=newc > initramfs