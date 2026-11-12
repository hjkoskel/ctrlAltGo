/*
simple video for linux 2 (v4l2) demonstration

takes /dev/video1 if avail
then /dev/video1

bcm2835_v4l2 is broken...

libcamera is the way but it have horrible dependencies

This is very crude, just demonstrating driver loading etc.. puhveri library is quite slow
*/
package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/hjkoskel/ctrlaltgo/puhveri"

	"github.com/hjkoskel/ctrlaltgo"
	"github.com/hjkoskel/ctrlaltgo/initializing"
	"github.com/peterhagelund/go-v4l2/v4l2"
	"golang.org/x/sys/unix"
)

func getIntEnvValue(name string, defaultvalue int64) int64 {
	s := os.Getenv("FOO")
	if s == "" {
		return defaultvalue
	}
	i, errParse := strconv.ParseInt(s, 10, 64)
	if errParse != nil {
		return defaultvalue
	}
	return i
}

func FileExists(fname string) bool {
	file, err := os.Open(fname)
	defer file.Close()
	return !errors.Is(err, os.ErrNotExist)
}

func yuvToRGB(Y, U, V int) color.Color {
	C := Y - 16
	D := U
	E := V

	R := clamp((298*C + 409*E + 128) >> 8)
	G := clamp((298*C - 100*D - 208*E + 128) >> 8)
	B := clamp((298*C + 516*D + 128) >> 8)

	return color.RGBA{uint8(R), uint8(G), uint8(B), 255}
}

func clamp(value int) int {
	if value < 0 {
		return 0
	} else if value > 255 {
		return 255
	}
	return value
}

func drawYUVtoPuhveri(target puhveri.PuhveriImage, yuyv []byte, width, height int) {
	idx := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x += 2 {
			Y1 := int(yuyv[idx])
			U := int(yuyv[idx+1]) - 128
			Y2 := int(yuyv[idx+2])
			V := int(yuyv[idx+3]) - 128
			idx += 4

			target.Set(x, y, yuvToRGB(Y1, U, V))
			target.Set(x+1, y, yuvToRGB(Y2, U, V))
		}
	}
}

const (
	KERNELDRIVERDIR = "/bootcard/modules/"
	MNT_BOOTCARD    = "/bootcard"
)

func mountBootPartition() error {

	bootPartitionName, errBootPartition := initializing.GetBootPartitionName([]string{"mmcblk0p1", "sda1"}, time.Minute) //USB enumeration takes time
	if errBootPartition != nil {
		return errBootPartition
	}

	if len(bootPartitionName) == 0 {
		fmt.Printf("Must be qemu without drive\n") //TODO detect when on qemu?
		os.MkdirAll(MNT_BOOTCARD, 0777)
		return nil
	}
	fmt.Printf("Using %s as boot partition\n", bootPartitionName)
	return initializing.CreateAndMount([]initializing.MountCmd{
		initializing.MountCmd{
			Source: "/dev/" + bootPartitionName,
			Target: MNT_BOOTCARD,
			FsType: "vfat",
			//Flags  uintptr
			//Data   string
		},
	})
}

func loadkernelModule(name string) error {
	modulePath := path.Join(KERNELDRIVERDIR, name)
	byt, errRead := os.ReadFile(modulePath)
	if errRead != nil {
		return fmt.Errorf("file %s load fail err:%s", modulePath, errRead)
	}
	errInit := unix.InitModule(byt, "")
	if errInit != nil {
		return fmt.Errorf("InitModule %s err:%s", modulePath, errInit)
	}
	return nil
}

// TODO problem with raspi webcam driver... now only usb webcam
func loadModulesForUsbWebcam() {
	names := []string{
		"mc.ko", //The least dependent module
		"uvc.ko",
		"videodev.ko",
		"videobuf2-common.ko",
		"videobuf2-v4l2.ko",
		"videobuf2-memops.ko",
		"videobuf2-vmalloc.ko",
		"uvcvideo.ko", //The most dependent module
	}
	for _, name := range names {
		fmt.Printf("loading module %s\n", name)
		errLoad := loadkernelModule(name)
		if errLoad != nil {
			fmt.Printf("FYI load %s error:%s\n", name, errLoad)
		}
	}
}

func main() {
	if os.Getpid() == 1 {
		//Need to initialize stuff, acts as init system
		errMount := initializing.MountNormal()
		ctrlaltgo.JamIfErr(errMount)
		errMountBoot := mountBootPartition()
		ctrlaltgo.JamIfErr(errMountBoot)
	}

	loadModulesForUsbWebcam() //Even when using prog manager, make sure that modules are loaded

	camXres := getIntEnvValue("CAMXRES", 640)
	camYres := getIntEnvValue("CAMYRES", 480)
	camname := os.Getenv("CAMNAME")

	if camname == "" {
		camname = "/dev/video0"
	}
	if !FileExists(camname) {
		camname = "/dev/video1"
	}

	if !FileExists(camname) {
		fmt.Printf("no video0 or video1\n")
		return
	}

	camcfg := v4l2.CameraConfig{
		Path:      camname,
		BufType:   v4l2.BufTypeVideoCapture,
		PixFormat: v4l2.PixFmtYUYV, // PixFmtMJPEG,
		Width:     uint32(camXres),
		Height:    uint32(camYres),
		Memory:    v4l2.MemoryMmap,
		BufCount:  1,
	}
	fmt.Printf("camcfg:%#v\n")
	camera, errCamera := v4l2.NewCamera(&camcfg)
	if errCamera != nil {
		fmt.Printf("error camera %s\n", errCamera)
		return
	}
	fmt.Printf("Driver....: %s\n", camera.Driver())
	fmt.Printf("Card......: %s\n", camera.Card())
	fmt.Printf("BusInfo...: %s\n", camera.BusInfo())
	errStream := camera.StreamOn()
	if errStream != nil {
		fmt.Printf("error stream %s\n", errStream)
		return
	}

	/*

	*********/
	fmt.Printf("stream is on, going to video mode\n")

	lastFbDevName, errName := puhveri.GetLastFbFileName()
	if errName != nil {
		ctrlaltgo.JamIfErr(fmt.Errorf("failed getting FB name %s", errName))
	}
	fmt.Printf("last fb name =%s\n", lastFbDevName)
	dev, devErr := puhveri.OpenFbPuhveri(lastFbDevName)
	if devErr != nil {
		ctrlaltgo.JamIfErr(fmt.Errorf("opening fb device failed %s", devErr))
	}
	fmt.Printf("change to console graphics\n")
	errConsole := puhveri.ConsoleToGraphics()
	if errConsole != nil {
		ctrlaltgo.JamIfErr(fmt.Errorf("error chancing to console: %s", errConsole))
	}

	imgBuffer, errImgBuffer := dev.GetBuffer()
	if errImgBuffer != nil {
		ctrlaltgo.JamIfErr(fmt.Errorf("error getting framebuffer:%s", errImgBuffer))
	}

	for {
		tFrameStart := time.Now()
		frame, errGrab := camera.GrabFrame()
		if errGrab != nil {
			fmt.Printf("error grabbing frame %s\n", errGrab)
			return
		}
		durGrap := time.Since(tFrameStart)

		drawYUVtoPuhveri(imgBuffer, frame, int(camXres), int(camYres))

		dev.UpdateScreenAreas(imgBuffer, []image.Rectangle{})

		fps := 1 / time.Since(tFrameStart).Seconds()
		fmt.Printf("FPS: %.1f grab:%s\n", fps, durGrap)

	}
}

/*
oading module mc.ko
FYI load mc.ko error:file /bootcard/modules/mc.ko load fail err:open /bootcard/modules/mc.ko: no such file or directory
loading module uvc.ko
FYI load uvc.ko error:file /bootcard/modules/uvc.ko load fail err:open /bootcard/modules/uvc.ko: no such file or directory
loading module videodev.ko
FYI load videodev.ko error:file /bootcard/modules/videodev.ko load fail err:open /bootcard/modules/videodev.ko: no such file or directory
loading module videobuf2-common.ko
FYI load videobuf2-common.ko error:file /bootcard/modules/videobuf2-common.ko load fail err:open /bootcard/modules/videobuf2-common.ko: no such file or directory
loading module videobuf2-v4l2.ko
FYI load videobuf2-v4l2.ko error:file /bootcard/modules/videobuf2-v4l2.ko load fail err:open /bootcard/modules/videobuf2-v4l2.ko: no such file or directory
loading module videobuf2-memops.ko
FYI load videobuf2-memops.ko error:file /bootcard/modules/videobuf2-memops.ko load fail err:open /bootcard/modules/videobuf2-memops.ko: no such file or directory
loading module videobuf2-vmalloc.ko
FYI load videobuf2-vmalloc.ko error:file /bootcard/modules/videobuf2-vmalloc.ko load fail err:open /bootcard/modules/videobuf2-vmalloc.ko: no such file or directory
loading module uvcvideo.ko
FYI load uvcvideo.ko error:file /bootcard/modules/uvcvideo.ko load fail err:open /bootcard/modules/uvcvideo.ko: no such file or directory


----

mc.ko   /lib/modules/6.12.47+rpt-rpi-v8/kernel/drivers/media/mc/mc.ko.xz
uvc.ko  /lib/modules/6.12.47+rpt-rpi-v8/kernel/drivers/media/common/uvc.ko.xz
videodev.ko /lib/modules/6.12.47+rpt-rpi-v8/kernel/drivers/media/v4l2-core/videodev.ko.xz
videobuf2-common.ko  /lib/modules/6.12.47+rpt-rpi-v8/kernel/drivers/media/common/videobuf2/videobuf2-common.ko.xz
videobuf2-v4l2.ko  /lib/modules/6.12.47+rpt-rpi-v8/kernel/drivers/media/common/videobuf2/videobuf2-v4l2.ko.xz
videobuf2-memops.ko /lib/modules/6.12.47+rpt-rpi-v8/kernel/drivers/media/common/videobuf2/videobuf2-memops.ko.xz
videobuf2-vmalloc.ko  /lib/modules/6.12.47+rpt-rpi-v8/kernel/drivers/media/common/videobuf2/videobuf2-vmalloc.ko.xz
uvcvideo.ko  /lib/modules/6.12.47+rpt-rpi-v8/kernel/drivers/media/usb/uvc/uvcvideo.ko.xz


*/
