package main

import (
	"bytes"
	"flag"
	"fmt"
	"strconv"
	"strings"
)

const base32chars string = "0123456789bcdefghjkmnpqrstuvwxyz"

type position struct {
	Lat float64
	Lng float64
}

var (
	latitude  float64
	longitude float64
	precision int
)

func init() {
	flag.Float64Var(&latitude, "lat", 0, "纬度")
	flag.Float64Var(&longitude, "lng", 0, "经度")
	flag.IntVar(&precision, "prec", 4, "精度")
}
func main() {
	flag.Parse()
	fmt.Println(NewPosition(latitude, longitude).GeoHash(precision))
}

func NewPosition(latitude float64, longitude float64) *position {
	return &position{
		latitude,
		longitude,
	}
}

func (p *position) GeoHash(precision int) string {
	if precision%2 != 0 || precision <= 0 {
		panic("精度小于等于与0或者不为偶数")
	}
	n := precision / 2 * 5
	var (
		strBuf   strings.Builder
		strArray []string
	)
	buf := p.transToGeoHashBinaryBuffer(n)

	for i := 0; i < 2*n; {
		strBuf.WriteString(strconv.FormatUint(uint64(buf[i]), 2))
		i++
		if i%5 == 0 {
			strArray = append(strArray, strBuf.String())
			strBuf.Reset()
		}
	}
	for _, val := range strArray {
		idx, _ := strconv.ParseUint(val, 2, 64)
		strBuf.WriteString(string(base32chars[idx]))
	}
	return strBuf.String()
}
func (p *position) transToGeoHashBinaryBuffer(n int) []byte {
	var buf bytes.Buffer
	latChan, lngChan := make(chan []byte), make(chan []byte)
	go getBinaryBuffer(p.Lat, -90, 90, n, latChan)
	go getBinaryBuffer(p.Lng, -180, 180, n, lngChan)
	latBuf := <-latChan
	lngBuf := <-lngChan
	for i := 0; i < n; i++ {
		buf.WriteByte(lngBuf[i])
		buf.WriteByte(latBuf[i])
	}
	return buf.Bytes()
}
func getBinaryBuffer(f float64, left float64, right float64, n int, latChan chan []byte) {
	var (
		buf bytes.Buffer
		val byte
		l   = left
		r   = right
	)
	for i := n; i > 0; i-- {
		val, l, r = getOneBinary(f, l, r)
		buf.WriteByte(val)
	}
	latChan <- buf.Bytes()
}

func getOneBinary(f float64, left float64, right float64) (byte, float64, float64) {
	mid := (left + right) / 2
	if left <= f && f < mid {
		return 0, left, mid
	} else if mid <= f && f <= right {
		return 1, mid, right
	} else {
		panic("overflow")
	}
}
