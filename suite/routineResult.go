package suite

import (
	"fmt"
)

func (tr *routineResult) validate() (trr routineResult, isValid bool, err []error) {
	isValid = true

	trr = *tr
	trr.Suite = nil

	for _, key := range tr.Expected.Return {
		if tr.Code != tr.Expected.Code {
			logger.Log(tr.EntityTitle, tr.RoutineDescription, "Return code expected", WARNING, map[string]interface{}{"expected": tr.Expected.Code, "returned": tr.Code}, false)

			isValid = false
		}
		if content, thisExist := tr.Result[key]; !thisExist {
			logger.Log(tr.EntityTitle, tr.RoutineDescription, "Key not found", WARNING, map[string]interface{}{"expected": key}, false)

			isValid = false
			err = append(err, fmt.Errorf("Key %s not found in %+v", key, tr.Result))
		} else {
			for _, variable := range *tr.CaseVariables {
				if originVar, destinyVar := translateVariableTransfer(variable); originVar == key || originVar == "." {
					tr.Suite.saveVariable(destinyVar, content)
				}
			}
		}
	}

	return
}
