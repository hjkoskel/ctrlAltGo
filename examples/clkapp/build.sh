CGO_ENABLED=0 go build -o ./buildfs/program
cd buildfs
#TODO need copy init to buildfs like oledgadget ui
cp ../../../textmonitor/cmd/oledgadget/init init

find . -print0 | cpio --null --create --verbose --format=newc | gzip --best > ../initramfs