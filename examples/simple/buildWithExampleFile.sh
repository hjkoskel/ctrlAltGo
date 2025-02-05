# build with example file
rm -r -f buildfs
mkdir buildfs
mkdir ./buildfs/bin
CGO_ENABLED=0 go build -o ./buildfs/init
echo "this is example file" > ./buildfs/bin/example.txt
cd buildfs
find . -print0 | cpio --null --create --verbose --format=newc | gzip --best > ../initramfs


