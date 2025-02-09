#qemu-system-x86_64 -m 2G -kernel ../bzImage -initrd initramfs -serial pty -display sdl,gl=on -device qxl

#Ok RTC time
#qemu-system-x86_64 -kernel ../bzImage -initrd initramfs -device qxl -netdev user,id=mynet0,hostfwd=tcp::8080-:80 -device e1000,netdev=mynet0

#Set to bad
qemu-system-x86_64 -kernel ../../examples/bzImage -initrd initramfs -device qxl -netdev user,id=mynet0,hostfwd=tcp::4242-:4242 -device e1000,netdev=mynet0 -rtc base=2006-06-06
