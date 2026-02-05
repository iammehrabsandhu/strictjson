package strictjson

type Decoder struct {
	DisallowUnknownFields bool
	SuggestClosest        bool
}

type DecoderOption func(*Decoder)

func NewDecoder(opts ...DecoderOption) *Decoder {
	d := &Decoder{
		DisallowUnknownFields: true,
		SuggestClosest:        false,
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

func WithDisallowUnknownFields(disallow bool) DecoderOption {
	return func(d *Decoder) {
		d.DisallowUnknownFields = disallow
	}
}

func WithSuggestClosest(suggest bool) DecoderOption {
	return func(d *Decoder) {
		d.SuggestClosest = suggest
	}
}
