#qemu-system-x86_64 -m 2G -kernel ../bzImage -initrd initramfs -serial pty -display sdl,gl=on -device qxl
qemu-system-x86_64 -kernel ../bzImage -initrd initramfs -device qxl -hda testdrive.img