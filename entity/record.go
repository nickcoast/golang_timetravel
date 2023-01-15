package entity

import "fmt"

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

func (d *Record) DataVal(key string) (val *string) {
	if v, exists := d.Data[key]; exists {
		return &v
	}
	return nil
}
func (d *Record) SetID(id int) {
	fmt.Println("Record.SetID", id)
	d.ID = id
	fmt.Println("Record.SetID", d.ID)
}
