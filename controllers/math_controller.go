/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"github.com/Knetic/govaluate"
	"k8s.io/apimachinery/pkg/api/errors"
	"strconv"

	mathv1alpha1 "github.com/example/math-operator/api/v1alpha1"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// MathReconciler reconciles a Math object
type MathReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type Operation func(x int, y int) int

//+kubebuilder:rbac:groups=math.example.com,resources=maths,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=math.example.com,resources=maths/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=math.example.com,resources=maths/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// In this example, reconcile will read an equation and variables defined in the
// applied spec and set the status to the result of the equation
func (r *MathReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = r.Log.WithValues("math", req.NamespacedName)

	// load math object with current Spec and Status from current context
	math := mathv1alpha1.Math{}
	err := r.Get(ctx, req.NamespacedName, &math)
	if err != nil {
		if errors.IsNotFound(err) {
			return ctrl.Result{}, nil
		}
	}

	// after we return we want to make sure to update the status
	defer r.Status().Update(ctx, &math)

	// load the expression from the current spec
	expression, err := govaluate.NewEvaluableExpression(math.Spec.Expression)
	if err != nil {
		math.Status.Message = err.Error()
		return ctrl.Result{}, err
	}

	// parse the variables from the spec into a type govaluate can use
	formattedParameters, err := r.parseVariablesFromSpec(math.Spec)

	if err != nil {
		math.Status.Message = err.Error()
		return ctrl.Result{}, err
	}

	result, err := expression.Evaluate(formattedParameters)
	if err != nil {
		math.Status.Message = err.Error()
		return ctrl.Result{}, err
	}

	formattedResult := fmt.Sprintf("%v", result)
	math.Status.Result = formattedResult
	math.Status.Message = "OK"
	return ctrl.Result{}, nil
}

func (r *MathReconciler) parseVariablesFromSpec(spec mathv1alpha1.MathSpec) (map[string]interface{}, error) {
	variables := spec.Variables
	formattedParameters := make(map[string]interface{}, len(variables))
	for key, value := range variables {
		formattedValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, err
		}
		formattedParameters[key] = formattedValue
	}
	return formattedParameters, nil
}


// SetupWithManager sets up the controller with the Manager.
func (r *MathReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mathv1alpha1.Math{}).
		Complete(r)
}
