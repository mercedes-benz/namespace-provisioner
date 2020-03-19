// SPDX-License-Identifier: MIT

package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	apiErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	Log                 logr.Logger
	ConfigNamespaceName string
}

const namespaceConfigMapAnnotation = "namespace-provisioner.daimler-tss.com/config"
const namespaceSecretAnnotation = "namespace-provisioner.daimler-tss.com/secret"

func (r *NamespaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	_ = context.Background()
	logger := r.Log.WithValues("namespace", req.NamespacedName)
	logger.Info("Reconciling namespace")

	// Fetch the Namespace Instance
	namespaceInstance := &corev1.Namespace{}
	err := r.Client.Get(context.TODO(), req.NamespacedName, namespaceInstance)
	if err != nil {
		if apiErrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, err
	}

	namespaceName := namespaceInstance.Name
	if namespaceInstance.Status.Phase == corev1.NamespaceActive {

		var configs []string
		if namespaceAnnotationVal, ok := namespaceInstance.ObjectMeta.Annotations[namespaceConfigMapAnnotation]; ok {

			logger.Info(fmt.Sprintf("Handle event for namespace %s with annotation %s=%s", namespaceName, namespaceConfigMapAnnotation, namespaceAnnotationVal))
			configs, err = r.addConfigFromConfigMap(configs, namespaceAnnotationVal, namespaceInstance)
			if err != nil {
				return ctrl.Result{}, err
			}
		}

		if namespaceAnnotationVal, ok := namespaceInstance.ObjectMeta.Annotations[namespaceSecretAnnotation]; ok {

			logger.Info(fmt.Sprintf("Handle event for namespace %s with annotation %s=%s", namespaceName, namespaceSecretAnnotation, namespaceAnnotationVal))
			configs, err = r.addConfigFromSecret(configs, namespaceAnnotationVal, namespaceInstance)
			if err != nil {
				return ctrl.Result{}, err
			}

		}

		if len(configs) > 0 {
			return r.newResourcesFromConfig(configs, namespaceInstance)
		}
		logger.V(1).Info(fmt.Sprintf("Ignore event for namespace %s without annotation ", namespaceName))

	} else {
		logger.V(1).Info(fmt.Sprintf("Ignore event for namespace %s with status %s", namespaceName, namespaceInstance.Status.Phase))
	}
	return ctrl.Result{}, nil
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}

func (r *NamespaceReconciler) addConfigFromConfigMap(configs []string, configMapNames string, namespace *corev1.Namespace) ([]string, error) {
	logger := r.Log.WithValues("namespace", namespace.Name)
	// configMapNames could be comma separated
	for _, configMapName := range strings.Split(configMapNames, ",") {

		configMap := &corev1.ConfigMap{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{Name: configMapName, Namespace: r.ConfigNamespaceName}, configMap)

		if err != nil {
			logger.Error(err, fmt.Sprintf("Error getting ConfigMap %s in namespace %s: %s", configMapName, r.ConfigNamespaceName, err))
			return configs, err
		}

		logger.V(1).Info(fmt.Sprintf("Found ConfigMap %s in namespace %s", configMapName, r.ConfigNamespaceName))
		for key, value := range configMap.Data {
			logger.V(1).Info(fmt.Sprintf("Add %s from ConfigMap %s in namespace %s", key, configMapName, r.ConfigNamespaceName))
			configs = append(configs, value)
		}
	}
	return configs, nil
}

func (r *NamespaceReconciler) addConfigFromSecret(configs []string, secretNames string, namespace *corev1.Namespace) ([]string, error) {
	logger := r.Log.WithValues("namespace", namespace.Name)
	// secrets could be comma separated
	for _, secretName := range strings.Split(secretNames, ",") {

		secret := &corev1.Secret{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{Name: secretName, Namespace: r.ConfigNamespaceName}, secret)

		if err != nil {
			logger.Error(err, fmt.Sprintf("Error getting Secret %s in namespace %s: %s", secretName, r.ConfigNamespaceName, err))
			return configs, err
		}
		logger.V(1).Info(fmt.Sprintf("Found Secret %s in namespace %s", secretName, r.ConfigNamespaceName))
		for key, value := range secret.Data {
			logger.V(1).Info(fmt.Sprintf("Add %s from Secret %s in namespace %s", key, secretName, r.ConfigNamespaceName))
			stringData := string(value)
			configs = append(configs, stringData)
		}

	}
	return configs, nil
}

func (r *NamespaceReconciler) newResourcesFromConfig(configs []string, namespace *corev1.Namespace) (ctrl.Result, error) {
	logger := r.Log.WithValues("namespace", namespace.Name)
	for _, config := range configs {

		var resources = strings.Split(config, "---")
		for _, resource := range resources {

			obj, groupKindVersion, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(resource), nil, nil)

			if err != nil {
				logger.Error(err, fmt.Sprintf("Error while decoding YAML object. Reason=%v", err))
				return ctrl.Result{}, err
			}

			if k8sObj, ok := obj.(runtime.Object); ok {

				accessor := meta.NewAccessor()
				accessor.SetNamespace(k8sObj, namespace.Name)

				name, err := accessor.Name(k8sObj)
				if err != nil {
					logger.Error(err, fmt.Sprintf("Error getting name of object %v", k8sObj))
					return ctrl.Result{}, err
				}

				// Check if this object already exists
				var found unstructured.Unstructured
				found.SetGroupVersionKind(k8sObj.GetObjectKind().GroupVersionKind())
				err = r.Client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace.Name}, &found)

				if err != nil && apiErrors.IsNotFound(err) {
					logger.Info(fmt.Sprintf("Try to create object of kind %s with version %s on namespace %s", groupKindVersion.Kind, groupKindVersion, namespace.Namespace))
					err = r.Client.Create(context.TODO(), k8sObj)
					if err != nil {
						return ctrl.Result{}, err
					}
				} else {
					logger.V(1).Info(fmt.Sprintf("Object %s already exists ", name))
				}
			}
		}
	}
	return ctrl.Result{}, nil
}
