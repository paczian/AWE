package cwl

import (
//"fmt"
//"reflect"
)

type CWL_object interface {
	CWL_minimal_interface
	GetClass() string
	GetId() string
	SetId(string)
	//is_Any()
}

type CWL_object_Impl struct {
	Id    string `yaml:"id,omitempty" json:"id,omitempty" bson:"id,omitempty"`
	Class string `yaml:"class,omitempty" json:"class,omitempty" bson:"class,omitempty"`
}

func (c *CWL_object_Impl) GetId() string   { return c.Id }
func (c *CWL_object_Impl) SetId(id string) { c.Id = id }

func (c *CWL_object_Impl) GetClass() string { return c.Class }

type CWL_object_array []CWL_object
