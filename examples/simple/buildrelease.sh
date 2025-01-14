CGO_ENABLED=0 go build -ldflags="-s -w" -o init
echo init | cpio -o --format=newc > initramfs


