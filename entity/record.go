package entity

type Record struct {
	ID   int               `json:"id"`
	Data map[string]string `json:"data"`
}

func (d *Record) Copy() Record {
	values := d.Data

	newMap := map[string]string{}
	for key, value := range values {
		newMap[key] = value
	}

	return Record{
		ID:   d.ID,
		Data: newMap,
	}
}

func (d *Record) DataVal(key string) (val string, err error) {
	if v, exists := d.Data[key]; exists {
		return v, nil
	}
	return "", Errorf("No such key '%v'", key)
}
