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
	mathv1beta1 "github.com/example/math-operator/api/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"path/filepath"
	"reflect"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"testing"
	"time"

	mathv1alpha1 "github.com/example/math-operator/api/v1alpha1"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = mathv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = mathv1beta1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&MathReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
		Log:    ctrl.Log.WithName("controllers").WithName("Math"),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		err = k8sManager.Start(ctrl.SetupSignalHandler())
		Expect(err).ToNot(HaveOccurred())
	}()

}, 60)

var _ = Describe("Math Controller", func() {
	const (
		MathNamespace = "default"
		MathName      = "mathtest"

		timeout  = time.Second * 10
		duration = time.Second * 10
		interval = time.Millisecond * 250
	)

	Context("When Updating Controller Spec", func() {
		It("Should set Status.Result to the result of the specified equation", func() {
			By("Doing Math")
			ctx := context.Background()
			mathJob := &mathv1beta1.Math{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "math.example.com/v1",
					Kind:       "Math",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      MathName,
					Namespace: MathNamespace,
				},
				Spec: mathv1beta1.MathSpec{
					Expression: "3 * x + 7",
					Variables: map[string]mathv1beta1.Variable{
						"x": {
							Type:  "float",
							Value: "4",
						},
					},
				},
			}
			Expect(k8sClient.Create(ctx, mathJob)).Should(Succeed())

			mathLookupKey := types.NamespacedName{Name: MathName, Namespace: MathNamespace}
			createdMath := &mathv1beta1.Math{}

			Eventually(func() bool {
				err := k8sClient.Get(ctx, mathLookupKey, createdMath)
				if err != nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

			Expect(createdMath.Spec.Expression).Should(Equal("3 * x + 7"))
			//Expect(createdMath.Spec.Variables).Should(Equal(map[string]string{"x": "4"}))

			testMath := &mathv1beta1.Math{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "testmathname",
					Namespace: MathNamespace,
				},
				Spec: mathv1beta1.MathSpec{
					Expression: "3 * x + 7",
					Variables: map[string]mathv1beta1.Variable{
						"x": {
							Type:  "float",
							Value: "4",
						},
					},
				},
			}

			kind := reflect.TypeOf(mathv1beta1.Math{}).Name()
			gvk := mathv1beta1.GroupVersion.WithKind(kind)

			controllerRef := metav1.NewControllerRef(createdMath, gvk)
			testMath.SetOwnerReferences([]metav1.OwnerReference{*controllerRef})
			Expect(k8sClient.Create(ctx, testMath)).Should(Succeed())

			By("By checking the Math have the correct result")
			Eventually(func() (string, error) {
				err := k8sClient.Get(ctx, mathLookupKey, createdMath)
				if err != nil {
					return "err", err
				}
				return createdMath.Status.Message, nil
			}, duration, interval).Should(Equal("OK"))
			Eventually(func() (string, error) {
				err := k8sClient.Get(ctx, mathLookupKey, createdMath)
				if err != nil {
					return "err", err
				}
				return createdMath.Status.Result, nil
			}, duration, interval).Should(Equal("19"))
		})
	})

})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
