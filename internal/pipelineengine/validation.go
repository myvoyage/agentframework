package pipelineengine

import (
	"encoding/json"
	"github.com/xeipuuv/gojsonschema"
	errutil "AgentFramework/internal/errors"
)

// ValidateInputs validates the inputs of a tool according to its ToolSpec.InputsSchema
func ValidateInputs(spec ToolSpec, params map[string]interface{}) error {
	if spec.InputsSchema == nil || len(spec.InputsSchema) == 0 {
		return nil
	}
	schemaBytes, err := json.Marshal(spec.InputsSchema)
	if err != nil {
		return errutil.NewAppError(errutil.ErrCodeInternal, "failed to marshal inputs schema", err)
	}
	docBytes, _ := json.Marshal(params)
	loaderSchema := gojsonschema.NewBytesLoader(schemaBytes)
	loaderDoc := gojsonschema.NewBytesLoader(docBytes)
	result, err := gojsonschema.Validate(loaderSchema, loaderDoc)
	if err != nil {
		return errutil.NewAppError(errutil.ErrCodeInternal, "failed to validate inputs", err)
	}
	if !result.Valid() {
		details := make([]string, 0, len(result.Errors()))
		for _, e := range result.Errors() {
			details = append(details, e.String())
		}
		return errutil.NewAppError(errutil.ErrCodeSchemaValidation, "input parameters validation failed", details)
	}
	return nil
}

// ValidateOutputs validates the outputs of a tool according to its ToolSpec.OutputsSchema
func ValidateOutputs(spec ToolSpec, outputs map[string]interface{}) error {
	if spec.OutputsSchema == nil || len(spec.OutputsSchema) == 0 {
		return nil
	}
	schemaBytes, err := json.Marshal(spec.OutputsSchema)
	if err != nil {
		return errutil.NewAppError(errutil.ErrCodeInternal, "failed to marshal outputs schema", err)
	}
	docBytes, _ := json.Marshal(outputs)
	loaderSchema := gojsonschema.NewBytesLoader(schemaBytes)
	loaderDoc := gojsonschema.NewBytesLoader(docBytes)
	result, err := gojsonschema.Validate(loaderSchema, loaderDoc)
	if err != nil {
		return errutil.NewAppError(errutil.ErrCodeInternal, "failed to validate inputs", err)
	}
	if !result.Valid() {
		details := make([]string, 0, len(result.Errors()))
		for _, e := range result.Errors() {
			details = append(details, e.String())
		}
		return errutil.NewAppError(errutil.ErrCodeSchemaValidation, "output validation failed", details)
	}
	return nil
}
