# Runllamafile

This example demonstrates how to run llama.cpp on minimal linux system, whyle golang handling all housekeeping required for running C++ binary.

llma.cpp is dynamically linked software. That causes some portability issues. There are no straightforward ways to build software as static

First idea was to just execute llamafile. In reality llamafiles are not so portable.
Running llamafile requires using shell.
https://github.com/Mozilla-Ocho/llamafile
llamafile is using cosmopolitan standard library.

Another idea 

Things like [appimage](https://appimage.org) require quite much stuff on runtime system like fuse etc..
https://appimage.org/

So it is better idea to start from basics.
- Check what libraries are in use with ldd
- Pack those and deliver to runtime environment
- Run executable by executing *ld-linux-* binary


# building

First check what shared object files are required by binary
~~~sh
ldd llama-server
~~~

output is something like this
~~~txt
linux-vdso.so.1 (0x0000795e5b613000)
libstdc++.so.6 => /usr/lib/libstdc++.so.6 (0x0000795e5ac00000)
libm.so.6 => /usr/lib/libm.so.6 (0x0000795e5b4da000)
libgomp.so.1 => /usr/lib/libgomp.so.1 (0x0000795e5b487000)
libgcc_s.so.1 => /usr/lib/libgcc_s.so.1 (0x0000795e5b459000)
libc.so.6 => /usr/lib/libc.so.6 (0x0000795e5aa0e000)
/lib64/ld-linux-x86-64.so.2 => /usr/lib64/ld-linux-x86-64.so.2 (0x0000795e5b615000)
~~~

on raspberry it looks like this
~~~txt
linux-vdso.so.1 (0x0000007faae70000)
libllama.so => /home/henri/llama.cpp/build/bin/libllama.so (0x0000007faa9b0000)
libggml.so => /home/henri/llama.cpp/build/bin/libggml.so (0x0000007faa980000)
libggml-base.so => /home/henri/llama.cpp/build/bin/libggml-base.so (0x0000007faa8a0000)
libstdc++.so.6 => /lib/aarch64-linux-gnu/libstdc++.so.6 (0x0000007faa680000)
libm.so.6 => /lib/aarch64-linux-gnu/libm.so.6 (0x0000007faa5e0000)
libgcc_s.so.1 => /lib/aarch64-linux-gnu/libgcc_s.so.1 (0x0000007faa5a0000)
libc.so.6 => /lib/aarch64-linux-gnu/libc.so.6 (0x0000007faa3f0000)
libggml-cpu.so => /home/henri/llama.cpp/build/bin/libggml-cpu.so (0x0000007faa330000)
/lib/ld-linux-aarch64.so.1 (0x0000007faae33000)
libgomp.so.1 => /lib/aarch64-linux-gnu/libgomp.so.1 (0x0000007faa2c0000)
~~~

One strategy is to re-create folder structure *usr/bin*, *usr/lib*, *lib/aarch64-linux-gnu* and transfer that to runtime environment.
And then execute dynamic loader or like *ld-linux-aarch64.so.1* or *ld-linux-x86-64.so.2*

if all dependency files have unique names and there is no special plans to share shared libraries
then all these *.so* files can be conviently placed under one directory and compressed on that dir with 
~~~sh
tar -czvf content.tar.gz *
~~~
move content.tar.gz to source folder 

Then also 
https://huggingface.co/unsloth/SmolLM2-135M-Instruct-GGUF/tree/main

dowload SmolLM2-135M-Instruct-Q4_K_M.gguf and save that to source code

and run 
~~~sh
CGO_ENABLED go build -o init
or if cross compiling to arm
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -o init
~~~

content.tar.gz is embedded inside executable as well as .gguf file

And  then create initramfs
~~~sh
echo init | cpio -o --format=newc > initramfs
~~~


# run on qemu

This program can be executed as standalone initramfs program or binary loaded with for example with **intraberry**

~~~sh
qemu-system-x86_64 -cpu host -enable-kvm -m 8G \
-kernel ../../examples/gfx_bzImage \
-netdev user,id=mynet0,hostfwd=tcp::4242-:4242,hostfwd=tcp::8888-:8888 \
-device e1000,netdev=mynet0 \
-initrd initramfs -serial pty \
-display sdl,gl=on -device qxl
~~~

it is important to run to use *-cpu host* option to make sure that CPU have same instruction set where llama-server is built.


# TODOs

In reality it is much better idea to have .gguf model file saved on block device and get that mounted on start.
But goal of this code is to just demonstrate howto run C/C++ programs