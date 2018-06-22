package databaseGrpc

// EmptyUnmarshaler will just hold the data to be unmarshaled later
type EmptyUnmarshaler struct {
	Data []byte
}

func (e *EmptyUnmarshaler) UnmarshalBinary(data []byte) (err error) {
	e.UnmarshalBinaryData(data)
	return
}

func (e *EmptyUnmarshaler) UnmarshalBinaryData(data []byte) (newData []byte, err error) {
	e.Data = data
	return []byte{}, nil
}

func (e *EmptyUnmarshaler) MarshalBinary() (rval []byte, err error) {
	return e.Data, nil
}
