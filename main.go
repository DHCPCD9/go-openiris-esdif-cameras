package main

import (
	"bytes"
	"flag"
	"fmt"
	"image/jpeg"
	"log"
	"net/http"
	"strconv"
	"syscall"
	"time"

	"github.com/google/gousb"
	"github.com/hybridgroup/mjpeg"
	"github.com/kevmo314/go-uvc"
	"github.com/kevmo314/go-uvc/pkg/descriptors"
)

// Defaults for UVC OpenIris EDIF
var VID = 0x303a
var PID = 0x8000

var port = flag.Int("port", 8000, "Port to output frames to.")

func getdevices() []*gousb.Device {
	ctx := gousb.NewContext()
	defer ctx.Close()

	devices, err := ctx.OpenDevices(func(desc *gousb.DeviceDesc) bool {
		return desc.Vendor == gousb.ID(VID) && desc.Product == gousb.ID(PID)
	})

	if err != nil {
		log.Println(err)
		log.Println("Error while obtaining devices.")
		return []*gousb.Device{}
	}

	for _, device := range devices {
		defer device.Close()
	}

	return devices
}

func main() {
	flag.Parse()
	devices := getdevices()

	if len(devices) == 0 {
		log.Fatalln("No devices found.")
	}

	mux := http.NewServeMux()
	for _, device := range devices {
		stream := mjpeg.NewLiveStream()
		stream.FrameInterval = 16600 * time.Microsecond
		deviceAddress := fmt.Sprintf("/dev/bus/usb/%03v/%03v", device.Desc.Bus, device.Desc.Address)

		go imagestreamer(stream, deviceAddress)
		mux.Handle(fmt.Sprintf("/stream/%d", device.Desc.Address), stream)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		response := "Available streams:\n"

		for _, device := range devices {
			response += fmt.Sprintf("/stream/%d\n", device.Desc.Address)
		}

		w.Write([]byte(response))
	})

	log.Printf("Streaming %d streams to localhost:%s", len(devices), strconv.Itoa(*port))
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":"+strconv.Itoa(*port)), mux))
}

func imagestreamer(stream *mjpeg.Stream, device string) {
frame:
	fd, err := syscall.Open(device, syscall.O_RDWR, 0)
	var deviceFd = fd
	if err != nil {
		panic(err)
	}
	ctx, err := uvc.NewUVCDevice(uintptr(fd))
	if err != nil {
		panic(err)
	}

	info, err := ctx.DeviceInfo()
	if err != nil {
		panic(err)
	}

	for _, iface := range info.StreamingInterfaces {

		for i, desc := range iface.Descriptors {
			fd, ok := desc.(*descriptors.MJPEGFormatDescriptor)
			if !ok {
				continue
			}
			frd := iface.Descriptors[i+1].(*descriptors.MJPEGFrameDescriptor)

			resp, err := iface.ClaimFrameReader(fd.Index(), frd.Index())

			if err != nil {
				log.Print("Yes")
				panic(err)
			}
			for {
				fr, err := resp.ReadFrame()
				if err != nil {
					log.Print(err)
					log.Print("Reclaiming Frame Reader and continuing to get frames... ")
					syscall.Close(deviceFd)
					goto frame
				}

				img, err := jpeg.Decode(fr)
				if err != nil {
					continue
				}
				jpegbuf := new(bytes.Buffer)

				if err = jpeg.Encode(jpegbuf, img, nil); err != nil {
					log.Printf("failed to encode: %v", err)
				}
				// boundry := ("--frame-boundary\r\nContent-Type: image/jpeg\r\nContent-Length: " + strconv.Itoa(len(jpegbuf.Bytes())) + "\r\n\r\n")
				// stream.UpdateJPEG(append([]byte(boundry), jpegbuf.Bytes()...))
				stream.UpdateJPEG(jpegbuf.Bytes())
			}
		}
	}
}
