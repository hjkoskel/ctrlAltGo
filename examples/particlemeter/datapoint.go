/*
datapoint
*/
package main

import (
	"time"

	"github.com/hjkoskel/splurts"
)

type ParticleMeas struct {
	BootNumber int32
	Uptime     int64
	Epoch      int64

	//Fill up latest information on ambient also
	Temperature float64 `splurts:"step=0.1,min=-40,max=40"`
	Humidity    float64 `splurts:"step=0.05,min=0,max=100"`
	Pressure    float64 `splurts:"step=100,min=85000,max=110000"`
	Small       float64 `splurts:"step=0.1,min=0,max=300"`
	Large       float64 `splurts:"step=0.1,min=0,max=300,infpos=99999,infneg=-99999"`
}

var PwParticleMeasm splurts.PiecewiseFloats

func GetParticleMeasPiecewise() error {
	var err error
	PwParticleMeasm, err = splurts.GetPiecewisesFromStruct(ParticleMeas{})
	return err
}

type ParticleMeasArr []ParticleMeas

func (p *ParticleMeasArr) Insert(newitem ParticleMeas, capacity int) {
	over := capacity - len(*p) - 1
	if 0 < over {
		*p = (*p)[over:]
	}
	*p = append(*p, newitem)
}

// Get based on epoch, pic from memory
func (p *ParticleMeasArr) GetMetricsAfterEpoch(t0 int64, t1 int64) []ParticleMeas { //Very simple and naive for demo
	result := []ParticleMeas{}
	for _, v := range *p {
		if t0 <= v.Epoch && v.Epoch <= t1 {
			result = append(result, v)
		}
	}
	return result
}

// Get some datapoints. Based on Epoch. set timescale=time.Hour then 1.0 means one hour on X axis
func (p *ParticleMeasArr) GetVecEpoch(t0 int64, timescale time.Duration) []float64 {
	result := make([]float64, len(*p))
	for i, v := range *p {
		result[i] = float64(v.Epoch-t0) / float64(timescale.Milliseconds())
	}
	return result
}

func (p *ParticleMeasArr) GetVecUptime(ut0 int64, timescale time.Duration) []float64 {
	result := make([]float64, len(*p))
	for i, v := range *p {
		result[i] = float64(v.Epoch-ut0) / float64(timescale.Milliseconds())
	}
	return result
}

func (p *ParticleMeasArr) GetVecTemperature() []float64 {
	result := make([]float64, len(*p))
	for i, v := range *p {
		result[i] = v.Temperature
	}
	return result
}
func (p *ParticleMeasArr) GetVecHumidity() []float64 {
	result := make([]float64, len(*p))
	for i, v := range *p {
		result[i] = v.Humidity
	}
	return result
}
func (p *ParticleMeasArr) GetVecPressure() []float64 {
	result := make([]float64, len(*p))
	for i, v := range *p {
		result[i] = v.Pressure
	}
	return result
}
func (p *ParticleMeasArr) GetVecSmall() []float64 {
	result := make([]float64, len(*p))
	for i, v := range *p {
		result[i] = v.Small
	}
	return result
}
func (p *ParticleMeasArr) GetVecLarge() []float64 {
	result := make([]float64, len(*p))
	for i, v := range *p {
		result[i] = v.Large
	}
	return result
}
