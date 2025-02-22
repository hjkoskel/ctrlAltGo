qemu-system-x86_64 -m 2G -kernel ../gfx_bzImage -netdev user,id=mynet0,hostfwd=tcp::4242-:4242 -device e1000,netdev=mynet0 -initrd initramfs -serial pty -display sdl,gl=on -device qxl
