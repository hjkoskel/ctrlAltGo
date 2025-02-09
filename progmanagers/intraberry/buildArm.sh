# Builds native on same arch wher
CGO_ENABLED=0 GOARCH=arm64 go build -o init
echo init | cpio -o --format=newc > initramfs