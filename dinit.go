package dinit

func Init(fns ...interface{}) error {
	r := &resolver{}
	if err := r.init(fns); err != nil {
		return err
	}
	if err := r.resolve(); err != nil {
		return err
	}
	for _, fn := range r.fns {
		if err := r.validate(fn, nil); err != nil {
			return err
		}
	}
	return r.invoke()
}
