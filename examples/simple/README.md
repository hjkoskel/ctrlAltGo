# Simple example 

This simple program just for testing compile and deployment process.
And acting as start point of next projects

## Building initramfs

~~~ sh
GOOS=linux CGO_ENABLED=0 go build -o init
// or
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o init
~~~

Or if want save some bytes, add option -ldflags="-s -w" 

Then finally 
~~~ sh
echo init | cpio -o --format=newc > initramfs
~~~

## Kernel

This simple example does not require much from kernel setup.

Qemu have some features like TODO LIST that is good to have turned on

## Running

### Running on Qemu


### Running on 

EFI boot stub

TODO hardware

https://wiki.archlinux.org/title/EFI_boot_stub
https://www.youtube.com/watch?v=ywrSDLp926M&t=34s

https://www.rodsbooks.com/efi-bootloaders/efistub.html
https://superuser.com/questions/1716534/booting-a-custom-linux-kernel-from-usb-on-real-hardware
https://www.kernel.org/doc/html/v5.9/admin-guide/efi-stub.html

### Running on raspberry pi

The easiest way to get started and keep system up to date is to create sdcard  with *rpi-imager* and choose 
*Raspberry Pi OS (othere)* and under there *Raspberry Pi OS Lite (64-bit)*

Write sdcard and then mount and replace following files

and copy initramfs to sdcard

This method works also under windows.

TODO GENERATE CONFIG FILES (tool...)

cmdline.txt
~~~ txt
console=tty1
~~~
config.txt
~~~
initramfs initramfs
arm_64bit=1
disable_overscan=1
arm_boost=1
~~~


# Testing drives

~~~sh
qemu-img create -f qcow2 testdrive.img 100M
~~~

Lets mount and do something
~~~sh
sudo modprobe nbd max_part=8
sudo qemu-nbd --connect=/dev/nbd0 testdrive.img
~~~
fdisk /dev/nbd0 -l
