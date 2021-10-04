package util

type ValidatorEntry struct {
	Fn  func() bool
	Err error
}

type Validator []ValidatorEntry

func (v Validator) Validate() error {
	for _, v := range v {
		if !v.Fn() {
			return v.Err
		}
	}
	return nil
}

type Validatable interface {
	GetValidator() Validator
}

func Validate(v Validatable) error {
	return v.GetValidator().Validate()
}
