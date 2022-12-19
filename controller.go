package main

import (
	"context"
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// reconcileReplicaSet reconciles ReplicaSets
type reconcileReplicaSet struct {
	// client can be used to retrieve objects from the APIServer.
	client client.Client
}

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &reconcileReplicaSet{}

func (r *reconcileReplicaSet) Reconcile(ctx context.Context, request reconcile.Request) (reconcile.Result, error) {
	// set up a convenient log object so we don't have to type request over and over again
	log := log.FromContext(ctx)

	// Fetch the pod from the cache
	pod := &corev1.Pod{}
	err := r.client.Get(ctx, request.NamespacedName, pod)
	if errors.IsNotFound(err) {
		log.Error(nil, "Could not find Pods")
		return reconcile.Result{}, nil
	}

	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not fetch pods: %+v", err)
	}

	// if the Namespace is a system one, do nothing.
	if request.Namespace == "kube-system" || strings.Contains(request.Namespace, "vmware") || strings.Contains(request.Namespace, "kubernetes") {
		return reconcile.Result{}, nil
	}

	registry := strings.Split(pod.Spec.Containers[0].Image, "/")

	if registry[0] == "avhost" {
		log.Info("Found docker.io")
		pod.Spec.Containers[0].Image = "otherrepo" + pod.Spec.Containers[0].Image
	}

	// Print the ReplicaSet
	log.Info("Reconciling Pods", "container name", pod.Spec.Containers[0].Image)

	err = r.client.Update(ctx, pod)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not update Pod data: %+v", err)
	}

	return reconcile.Result{}, nil
}
