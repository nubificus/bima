package image

type ImageType uint8

const (
	Unikernel ImageType = iota
	IOT
)

func (s ImageType) String() string {
	switch s {
	case Unikernel:
		return "unikernel"
	case IOT:
		return "iot"
	}
	return "unknown"
}
