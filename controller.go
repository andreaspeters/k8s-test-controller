package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
	// Fetch the pod from the cache
	pod := &corev1.Pod{}
	err := r.client.Get(ctx, request.NamespacedName, pod)
	if err != nil {
		logrus.WithField("func", "main.Reconcile").Error("Could not find Pods")
		return reconcile.Result{}, fmt.Errorf("could not find pods: %s", err)
	}

	var ns bool
	ns, err = regexp.MatchString(EnvNamespaces, request.Namespace)
	if err != nil {
		logrus.WithField("func", "main.Reconcile").Error("could not find namespace: " + err.Error())
		return reconcile.Result{}, fmt.Errorf("could not find namespace: %s", err)
	}
	// Skip this namespace
	if ns {
		return reconcile.Result{}, nil
	}

	for i, container := range pod.Spec.Containers {
		var skip bool
		// do not replace this image repository
		skip, err = regexp.MatchString(EnvSkipImageRepo, request.Namespace)

		if err != nil && !skip {
			pod.Spec.Containers[i].Image = EnvImageRepo + container.Image
		}
	}

	// Print the ReplicaSet
	logrus.WithField("func", "main.Reconcile").Info("Reconciling Pods", "container name", pod.Spec.Containers[0].Image)

	err = r.client.Update(ctx, pod)
	if err != nil {
		return reconcile.Result{}, fmt.Errorf("could not write pod: %s", err)
	}

	return reconcile.Result{}, nil
}
