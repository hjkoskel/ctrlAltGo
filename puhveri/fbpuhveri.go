package puhveri

import (
	"errors"
	"fmt"
	"image"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

type FbPuhveri struct {
	fd    uintptr
	mmap  []byte
	finfo FixScreeninfo
	vinfo VarScreeninfo
	//visualRect image.Rectangle
	//stride     int  int(d.finfo.Line_length)
}

func (p *FbPuhveri) virtualRect() image.Rectangle {
	return image.Rect(0, 0, int(p.vinfo.Xres_virtual), int(p.vinfo.Yres_virtual))
}

func (p *FbPuhveri) visualRect() image.Rectangle {
	return image.Rect(int(p.vinfo.Xoffset), int(p.vinfo.Yoffset), int(p.vinfo.Xres), int(p.vinfo.Yres))
}

func GetLastFbFileName() (string, error) {
	for n := 0; n < 999; n++ {
		_, err := os.Stat(fmt.Sprintf("/dev/fb%v", n))
		if err != nil {
			if n == 0 {
				return "", fmt.Errorf("fb files not found")
			}
			return fmt.Sprintf("/dev/fb%v", n-1), nil
		}
	}
	return "", fmt.Errorf("too many fbs. Error?")
}

func OpenFbPuhveri(devFile string) (Puhveri, error) {
	fd, err := unix.Open(devFile, unix.O_RDWR|unix.O_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("open %s: %v", devFile, err)
	}
	if int(uintptr(fd)) != fd {
		unix.Close(fd)
		return nil, errors.New("fd overflows")
	}
	d := FbPuhveri{fd: uintptr(fd)}

	//TODO laittaa funktioksi
	_, _, eno := unix.Syscall(unix.SYS_IOCTL, d.fd, FBIOGET_FSCREENINFO, uintptr(unsafe.Pointer(&d.finfo)))
	if eno != 0 {
		unix.Close(fd)
		return nil, fmt.Errorf("FBIOGET_FSCREENINFO: %v", eno)
	}

	d.mmap, err = unix.Mmap(fd, 0, int(d.finfo.Smem_len), unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		unix.Close(fd)
		return nil, fmt.Errorf("mmap: %v", err)
	}
	//---------------------
	d.vinfo, err = d.VarScreeninfo()
	if err != nil {
		return nil, err
	}
	//d.bitsPerPixel = int(vinfo.Bits_per_pixel)

	virtual := d.virtualRect() //image.Rect(0, 0, int(d.vinfo.Xres_virtual), int(d.vinfo.Yres_virtual))
	virtualPixels := virtual.Dx() * virtual.Dy()

	if len(d.mmap) < virtualPixels*int(d.vinfo.Bits_per_pixel)/8 {
		return nil, errors.New("framebuffer is too small")
	}

	if !d.visualRect().In(virtual) {
		return nil, errors.New("visual resolution not contained in virtual resolution")
	}

	return &d, nil
}

func (p *FbPuhveri) VarScreeninfo() (VarScreeninfo, error) {
	var vinfo VarScreeninfo
	_, _, eno := unix.Syscall(unix.SYS_IOCTL, p.fd, FBIOGET_VSCREENINFO, uintptr(unsafe.Pointer(&vinfo)))
	if eno != 0 {
		return vinfo, fmt.Errorf("FBIOGET_VSCREENINFO: %v", eno)
	}
	return vinfo, nil
}

func (p *FbPuhveri) GetBuffer() (PuhveriImage, error) { //Just new image in correct format

	vr := p.visualRect()
	pixbuf := make([]byte, vr.Dx()*vr.Dy()*int(p.vinfo.Bits_per_pixel)/8)
	stride := int(p.finfo.Line_length)
	switch p.vinfo.Bits_per_pixel {
	case 32:
		return &ImageBGRA{Pix: pixbuf, Rect: vr, Stride: stride}, nil
	case 16:
		/*if vinfo.Grayscale == 1 {
			return &image.Gray16{Pix: pixbuf, Rect: vr, Stride: stride}, nil
		}*/
		return &ImageBGR565{Pix: pixbuf, Rect: vr, Stride: stride}, nil
	}

	return nil, fmt.Errorf("image with %v bits per pixel not supported", p.vinfo.Bits_per_pixel)
}

func (p *FbPuhveri) UpdateScreenAreas(buf PuhveriImage, areas []image.Rectangle) error { //Update to screen area. If not areas

	//if len(areas) == 0 {
	bArr := buf.GetRaw()

	/*for i, v := range bArr {
		p.mmap[i] = v
	}*/
	if len(areas) == 0 {
		copy(p.mmap, bArr)
		return nil
	}
	//}
	//TODO
	return nil
}

func (p *FbPuhveri) Close() error {
	e1 := unix.Munmap(p.mmap)
	if e2 := unix.Close(int(p.fd)); e2 != nil {
		return e2
	}
	return e1
}
