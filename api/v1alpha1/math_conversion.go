package v1alpha1

import (
	"github.com/example/math-operator/api/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
)

func (src *Math) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1beta1.Math)
	dst.Spec.Expression = src.Spec.Expression
	convertedVariables := map[string]v1beta1.Variable{}
	rawVariables := src.Spec.Variables
	for variable, value := range rawVariables {
		convertedVariables[variable] = v1beta1.Variable{
			Value: value,
			Type:  "float",
		}
	}
	dst.Spec.Variables = convertedVariables
	dst.Status.Result = src.Status.Result
	dst.Status.Message = src.Status.Message
	dst.ObjectMeta = src.ObjectMeta
	return nil
}

func (dst *Math) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1beta1.Math)
	dst.Spec.Expression = src.Spec.Expression
	convertedVariables := map[string]string{}
	rawVariables := src.Spec.Variables
	for variable, valuePair := range rawVariables {
		if valuePair.Type == "float" || valuePair.Type == "int" {
			convertedVariables[variable] = valuePair.Value
		} else {
			convertedVariables[variable] = "0"
		}
	}
	dst.Spec.Variables = convertedVariables
	dst.Status.Result = src.Status.Result
	dst.Status.Message = src.Status.Message
	dst.ObjectMeta = src.ObjectMeta
	return nil
}
