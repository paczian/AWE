package cwl

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/mapstructure"
	//"os"
	"reflect"
	//"strings"
	//"gopkg.in/mgo.v2/bson"
)

type Workflow struct {
	CWL_object_Impl `bson:",inline" json:",inline" mapstructure:",squash"` // provides Id and Class fields
	Inputs          []InputParameter                                       `yaml:"inputs,omitempty" bson:"inputs,omitempty" json:"inputs,omitempty"`
	Outputs         []WorkflowOutputParameter                              `yaml:"outputs,omitempty" bson:"outputs,omitempty" json:"outputs,omitempty"`
	Steps           []WorkflowStep                                         `yaml:"steps,omitempty" bson:"steps,omitempty" json:"steps,omitempty"`
	Requirements    []interface{}                                          `yaml:"requirements,omitempty" bson:"requirements,omitempty" json:"requirements,omitempty"` //[]Requirement
	Hints           []interface{}                                          `yaml:"hints,omitempty" bson:"hints,omitempty" json:"hints,omitempty"`                      // []Requirement TODO Hints may contain non-requirement objects. Give warning in those cases.
	Label           string                                                 `yaml:"label,omitempty" bson:"label,omitempty" json:"label,omitempty"`
	Doc             string                                                 `yaml:"doc,omitempty" bson:"doc,omitempty" json:"doc,omitempty"`
	CwlVersion      CWLVersion                                             `yaml:"cwlVersion,omitempty" bson:"cwlVersion,omitempty" json:"cwlVersion,omitempty"`
	Metadata        map[string]interface{}                                 `yaml:"metadata,omitempty" bson:"metadata,omitempty" json:"metadata,omitempty"`
}

func (w *Workflow) GetClass() string { return string(CWL_Workflow) }
func (w *Workflow) GetId() string    { return w.Id }
func (w *Workflow) SetId(id string)  { w.Id = id }
func (w *Workflow) Is_CWL_minimal()  {}
func (w *Workflow) Is_Any()          {}
func (w *Workflow) Is_process()      {}

func GetMapElement(m map[interface{}]interface{}, key string) (value interface{}, err error) {

	for k, v := range m {
		k_str, ok := k.(string)
		if ok {
			if k_str == key {
				value = v
				return
			}
		}
	}
	err = fmt.Errorf("Element \"%s\" not found in map", key)
	return
}

func NewWorkflow(original interface{}) (workflow_ptr *Workflow, err error) {

	// convert input map into input array

	original, err = MakeStringMap(original)
	if err != nil {
		err = fmt.Errorf("(NewWorkflow) MakeStringMap returned: %s", err.Error())
		return
	}

	workflow := Workflow{}
	workflow_ptr = &workflow

	switch original.(type) {
	case map[string]interface{}:
		object := original.(map[string]interface{})
		inputs, ok := object["inputs"]
		if ok {
			err, object["inputs"] = NewInputParameterArray(inputs)
			if err != nil {
				err = fmt.Errorf("(NewWorkflow) NewInputParameterArray returned: %s", err.Error())
				return
			}
		}

		outputs, ok := object["outputs"]
		if ok {
			object["outputs"], err = NewWorkflowOutputParameterArray(outputs)
			if err != nil {
				err = fmt.Errorf("(NewWorkflow) NewWorkflowOutputParameterArray returned: %s", err.Error())
				return
			}
		}

		// convert steps to array if it is a map
		steps, ok := object["steps"]
		if ok {
			err, object["steps"] = CreateWorkflowStepsArray(steps)
			if err != nil {
				err = fmt.Errorf("(NewWorkflow) CreateWorkflowStepsArray returned: %s", err.Error())
				return
			}
		}

		requirements, ok := object["requirements"]
		if ok {
			fmt.Println("---- Workflow (before CreateRequirementArray) ----")
			spew.Dump(object)
			object["requirements"], err = CreateRequirementArray(requirements)
			if err != nil {
				fmt.Println("---- Workflow ----")
				spew.Dump(object)
				fmt.Println("---- requirements ----")
				spew.Dump(requirements)
				err = fmt.Errorf("(NewWorkflow) CreateRequirementArray returned: %s", err.Error())
				return
			}
		}

		fmt.Printf("......WORKFLOW raw")
		spew.Dump(object)
		//fmt.Printf("-- Steps found ------------") // WorkflowStep
		//for _, step := range elem["steps"].([]interface{}) {

		//	spew.Dump(step)

		//}

		err = mapstructure.Decode(object, &workflow)
		if err != nil {
			err = fmt.Errorf("(NewWorkflow) error parsing workflow class: %s", err.Error())
			return
		}
		fmt.Printf(".....WORKFLOW")
		spew.Dump(workflow)
		return

	default:

		err = fmt.Errorf("(NewWorkflow) Input type %s can not be parsed", reflect.TypeOf(original))
		return
	}
	//switch object["requirements"].(type) {
	//case map[interface{}]interface{}:
	// Convert map of outputs into array of outputs
	//	object["requirements"], err = CreateRequirementArray(object["requirements"])
	//	if err != nil {
	//		return
	//	}
	//case []interface{}:
	//	req_array := []Requirement{}

	//	for _, requirement_if := range object["requirements"].([]interface{}) {
	//		switch requirement_if.(type) {

	//		case map[interface{}]interface{}:

	//			requirement_map_if := requirement_if.(map[interface{}]interface{})
	//			requirement_data_if, xerr := GetMapElement(requirement_map_if, "class")

	//			if xerr != nil {
	///				err = fmt.Errorf("Not sure how to parse Requirements, class not found")
	//				return
	//			}

	//			switch requirement_data_if.(type) {
	//			case string:
	//				requirement_name := requirement_data_if.(string)
	//				requirement, xerr := NewRequirement(requirement_name, requirement_data_if)
	//				if xerr != nil {
	//					err = fmt.Errorf("error creating Requirement %s: %s", requirement_name, xerr.Error())
	//					return
	//				}
	//				req_array = append(req_array, requirement)
	//			default:
	//				err = fmt.Errorf("Not sure how to parse Requirements, not a string")
	//				return

	//			}
	//		default:
	//			err = fmt.Errorf("Not sure how to parse Requirements, map expected")
	//			return

	//		} // end switch

	//	} // end for
	//
	//object["requirements"] = req_array
	//}
	return
}
