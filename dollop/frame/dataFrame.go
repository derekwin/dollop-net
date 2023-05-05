package frame

// client send RequestDataSreamFrame to apply a new stream from server
type DataFrame struct {
	data []byte
}

func (df DataFrame) Type() Type {
	return DataFrameTag
}

func (df DataFrame) Encode() []byte {
	return WriteFrame(df.data, DataFrameTag)
}

func (df DataFrame) GetData() []byte {
	return df.data
}

func NewDataFrame(data []byte) DataFrame {
	return DataFrame{data: data}
}

func ParseDataFrame(data []byte) (DataFrame, error) {
	return DataFrame{data: data}, nil
}
