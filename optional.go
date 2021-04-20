package gocsv

type OrElse func(err error) interface{}

type OptionalRowData interface {
	Exist() bool
	Value() interface{}
	Err() error
	OrElse(orElse OrElse) interface{}
}

type Row struct {
	V     interface{}
	Error error
}

func (d *Row) Exist() bool {
	return d.V != nil
}

func (d *Row) Value() interface{} {
	return d.V
}

func (d *Row) Err() error {
	return d.Error
}

func (d *Row) OrElse(f OrElse) interface{} {
	if d.Exist() {
		return d.Value()
	}
	return f(d.Error)
}
