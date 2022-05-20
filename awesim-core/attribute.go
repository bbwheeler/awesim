package main

type attribute struct {
	data string | float64 | int64
}

func (a attribute) String() string {
	return string(a.data)
}

func (a attribute) Number() float64 {
	return float64(a.data)
}

func (a attribute) Integer() int64 {
	return int64(a.data)
}