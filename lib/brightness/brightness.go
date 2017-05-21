package brightness

type Brightness struct {
	maxFile        string
	brightnessFile string
	maxValue       float64
}

func New() (*Brightness, error) {
	b := &Brightness{}
	if err := b.init(); err != nil {
		return nil, err
	}
	return b, nil
}
