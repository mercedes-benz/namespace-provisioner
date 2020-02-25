// SPDX-License-Identifier: MIT
package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/networking/v1beta1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var _ = Describe("NamespaceController", func() {

	var (
		configMapWithIngressSpec *corev1.ConfigMap
		namespaceToTest          *corev1.Namespace
		//s                        *runtime.Scheme
		request     reconcile.Request
		ingressName types.NamespacedName
	)

	BeforeEach(func() {

		// Set the logger to development mode for verbose logs.
		logf.SetLogger(logf.ZapLogger(true))

		// config map in the config-namespace, containing the spec for a ingress
		configMapWithIngressSpec = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "resources-to-deploy",
				Namespace: "config-namespace",
			},
			Data: map[string]string{
				"my-ingress": `
								{
								    "apiVersion": "networking.k8s.io/v1beta1",
								    "kind": "Ingress",
								    "metadata": {
								        "name": "test-ingress"
								    },
								    "spec": {
								        "backend": {
								            "serviceName": "testsvc",
								            "servicePort": 80
								        }
								    }
								}`,
			},
		}

		// the namespace to test.
		namespaceToTest = &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: "my-namespace",
			},
		}

		// Mock request to simulate Reconcile() being called on an event for a watched resource .
		request = reconcile.Request{
			NamespacedName: types.NamespacedName{
				Name: "my-namespace",
			},
		}

		// The name, which should be created
		ingressName = types.NamespacedName{
			Name:      "test-ingress",
			Namespace: "my-namespace",
		}

		// Register operator types with the runtime scheme.
		s := scheme.Scheme
		s.AddKnownTypes(corev1.SchemeGroupVersion, namespaceToTest)

	})

	Describe("Reconcile ", func() {
		Context("Active Namespace", func() {
			Context("With config annotation", func() {
				It("should deploy contents of configMap", func() {

					namespaceToTest.Annotations = map[string]string{
						"namespace-provisioner.daimler-tss.com/config": "resources-to-deploy",
					}
					namespaceToTest.Status = corev1.NamespaceStatus{Phase: "Active"}

					// Create a fake client to mock API calls.
					cl := fake.NewFakeClientWithScheme(scheme.Scheme, []runtime.Object{
						namespaceToTest,
						configMapWithIngressSpec,
					}...)

					// Create a ReconcileNamespace object with the scheme and fake client.
					r := &NamespaceReconciler{
						Client:              cl,
						Log:                 zap.New(zap.UseDevMode(true)),
						ConfigNamespaceName: "config-namespace",
					}

					res, err := r.Reconcile(request)
					Expect(err).To(BeNil())
					Expect(res.Requeue).To(BeFalse())

					// Check if deployment has been created and has the correct size.
					ingress := &v1beta1.Ingress{}
					err = cl.Get(context.TODO(), ingressName, ingress)
					Expect(err).To(BeNil())
					Expect(ingress.Name).To(Equal("test-ingress"))
				})
			})

			Context("Without config annotation", func() {
				It("should not handle namespace", func() {

					namespaceToTest.Status = corev1.NamespaceStatus{Phase: "Active"}

					// Create a fake client to mock API calls.
					cl := fake.NewFakeClientWithScheme(scheme.Scheme, []runtime.Object{
						namespaceToTest,
						configMapWithIngressSpec,
					}...)

					// Create a ReconcileNamespace object with the scheme and fake client.
					r := &NamespaceReconciler{
						Client:              cl,
						Log:                 zap.New(zap.UseDevMode(true)),
						ConfigNamespaceName: "config-namespace",
					}

					res, err := r.Reconcile(request)
					Expect(err).To(BeNil())
					Expect(res.Requeue).To(BeFalse())

					err = cl.Get(context.TODO(), ingressName, &v1beta1.Ingress{})
					Expect(err).NotTo(BeNil())
					Expect(apiErrors.IsNotFound(err)).To(BeTrue())
				})
			})
			Context("With config annotation to non-existing config map", func() {
				It("should not handle namespace", func() {

					namespaceToTest.Annotations = map[string]string{
						"namespace-provisioner.daimler-tss.com/config": "not-existing-config-map",
					}
					namespaceToTest.Status = corev1.NamespaceStatus{Phase: "Active"}

					// Create a fake client to mock API calls.
					cl := fake.NewFakeClientWithScheme(scheme.Scheme, []runtime.Object{
						namespaceToTest,
						configMapWithIngressSpec,
					}...)

					// Create a ReconcileNamespace object with the scheme and fake client.
					r := &NamespaceReconciler{
						Client:              cl,
						Log:                 zap.New(zap.UseDevMode(true)),
						ConfigNamespaceName: "config-namespace",
					}

					res, err := r.Reconcile(request)
					Expect(err).NotTo(BeNil())
					Expect(res.Requeue).To(BeFalse())

					err = cl.Get(context.TODO(), ingressName, &v1beta1.Ingress{})
					Expect(err).NotTo(BeNil())
					Expect(apiErrors.IsNotFound(err)).To(BeTrue())
				})
			})
		})

		Context("Terminating Namespace", func() {

			It("should not handle namespace", func() {

				namespaceToTest.Annotations = map[string]string{
					"namespace-provisioner.daimler-tss.com/config": "resources-to-deploy",
				}
				namespaceToTest.Status = corev1.NamespaceStatus{Phase: "Terminating"}

				// Create a fake client to mock API calls.
				cl := fake.NewFakeClientWithScheme(scheme.Scheme, []runtime.Object{
					namespaceToTest,
					configMapWithIngressSpec,
				}...)

				// Create a ReconcileNamespace object with the scheme and fake client.
				r := &NamespaceReconciler{
					Client:              cl,
					Log:                 zap.New(zap.UseDevMode(true)),
					ConfigNamespaceName: "config-namespace",
				}

				res, err := r.Reconcile(request)
				Expect(err).To(BeNil())
				Expect(res.Requeue).To(BeFalse())

				err = cl.Get(context.TODO(), ingressName, &v1beta1.Ingress{})
				Expect(err).NotTo(BeNil())
				Expect(apiErrors.IsNotFound(err)).To(BeTrue())
			})
		})
	})
})
