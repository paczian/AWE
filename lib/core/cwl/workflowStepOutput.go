package cwl

import (
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/mitchellh/mapstructure"
	"reflect"
)

type WorkflowStepOutput struct {
	Id string `yaml:"id" bson:"id" json:"id" mapstructure:"id"`
}

func NewWorkflowStepOutput(original interface{}) (wso_ptr *WorkflowStepOutput, err error) {

	original, err = MakeStringMap(original)
	if err != nil {
		return
	}

	var wso WorkflowStepOutput

	switch original.(type) {
	case string:
		original_str := original.(string)
		wso = WorkflowStepOutput{Id: original_str}
		wso_ptr = &wso
		return
	case map[string]interface{}:
		err = mapstructure.Decode(original, &wso)
		if err != nil {
			err = fmt.Errorf("(CreateWorkflowStepOutputArray) %s", err.Error())
			return
		}
		wso_ptr = &wso
		return
	default:
		err = fmt.Errorf("(NewWorkflowStepOutput) could not parse NewWorkflowStepOutput, type unknown %s", reflect.TypeOf(original))
		return
	}

	return
}

func NewWorkflowStepOutputArray(original interface{}) (new_array []WorkflowStepOutput, err error) {

	switch original.(type) {
	case map[interface{}]interface{}:

		for k, v := range original.(map[interface{}]interface{}) {
			//fmt.Printf("A")

			wso, xerr := NewWorkflowStepOutput(v)
			//var output_parameter WorkflowStepOutput
			//err = mapstructure.Decode(v, &output_parameter)
			if xerr != nil {
				err = fmt.Errorf("(CreateWorkflowStepOutputArray) %s", xerr.Error())
				return
			}

			wso.Id = k.(string)
			//fmt.Printf("C")
			new_array = append(new_array, *wso)
			//fmt.Printf("D")

		}

	case []interface{}:
		for _, v := range original.([]interface{}) {

			wso, xerr := NewWorkflowStepOutput(v)
			if xerr != nil {
				err = xerr
				return
			}
			new_array = append(new_array, *wso)

		}
	default:
		err = fmt.Errorf("(NewWorkflowStepOutputArray) could not parse NewWorkflowStepOutputArray, type unknown %s", reflect.TypeOf(original))
		return
	} // end switch

	//spew.Dump(new_array)
	return
}
