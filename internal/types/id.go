package types

type ID int64

func (id ID) Valid() bool {
	return id > 0
}
