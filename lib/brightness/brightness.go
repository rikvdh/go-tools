package brightness

// Brightness is the holding structure for brightness-control
type Brightness struct {
	maxFile        string
	brightnessFile string
	maxValue       float64
}

// New creates and initializes a Brightness option
func New() (*Brightness, error) {
	b := &Brightness{}
	if err := b.init(); err != nil {
		return nil, err
	}
	return b, nil
}
