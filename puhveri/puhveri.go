/*
framebuffer library, needs much optimization
*/
package puhveri

import (
	"image"
)

type Puhveri interface {
	//created by OpenNAME(dev string) (Puhveri, error)
	//For faster operations. Minimize color conversions Just get buffer. Flash update buffer to screen
	GetBuffer() (PuhveriImage, error)                                  //Just new image in correct format
	UpdateScreenAreas(buf PuhveriImage, areas []image.Rectangle) error //Update to screen area. If not areas
	Close() error
}
