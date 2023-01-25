/*
Copyright 2023.

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
	"k8s.io/apimachinery/pkg/api/errors"
	"kafka.samoxbalz.io/util"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	apiv1 "kafka.samoxbalz.io/api/v1"
)

// TopicReconciler reconciles a Topic object
type TopicReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

const topicFinalizer = "kafka.samoxbalz.io/finalizer"

//+kubebuilder:rbac:groups=kafka.samoxbalz.io.my.domain,resources=topics,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kafka.samoxbalz.io.my.domain,resources=topics/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kafka.samoxbalz.io.my.domain,resources=topics/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Topic object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.12.2/pkg/reconcile
func (r *TopicReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Start reconciler")

	topic := &apiv1.Topic{}
	err := r.Get(ctx, req.NamespacedName, topic)

	if err != nil {
		if errors.IsNotFound(err) {
			logger.Info("Topic resource " + req.Name + " not found or deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get Topic resource")
		return ctrl.Result{}, err
	}

	kafkaClient, errKafkaConn := util.InitKafkaConnect(&topic.Spec)

	if errKafkaConn != nil {
		logger.Error(err, "Failed to connect to Kafka")
		return ctrl.Result{}, errKafkaConn
	}

	defer func() {
		err := kafkaClient.Close()
		if err != nil {
			logger.Error(err, "Failed to close connection")
		}
	}()

	// Add finalizer
	if !controllerutil.ContainsFinalizer(topic, topicFinalizer) {
		controllerutil.AddFinalizer(topic, topicFinalizer)
		err := r.Update(ctx, topic)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	if controllerutil.ContainsFinalizer(topic, topicFinalizer) {
		if !topic.GetDeletionTimestamp().IsZero() {
			logger.Info("Topic resource " + topic.Name + " to be deleted")
			controllerutil.RemoveFinalizer(topic, topicFinalizer)
			if errDeletionTopic := kafkaClient.DeleteTopic(topic.Spec.Name); errDeletionTopic != nil {
				return ctrl.Result{}, errDeletionTopic
			}
			err := r.Update(ctx, topic)
			if err != nil {
				return ctrl.Result{}, err
			}
			return ctrl.Result{}, nil
		}
	}

	// Hash for idempotency
	actualHash := util.GetHashFromLabels(topic.Labels)
	expectedHash := util.GenerateHashFromSpec(topic.Spec)

	if actualHash == "" {
		logger.Info("Topic resource " + topic.Name + " created")
		_, err = r.updateLabels(ctx, topic, expectedHash)
	} else if actualHash != expectedHash {
		logger.Info("Topic resource " + topic.Name + " changed")
		_, err = r.updateLabels(ctx, topic, expectedHash)
	}

	return ctrl.Result{}, nil
}

func (r *TopicReconciler) updateLabels(ctx context.Context, topic *apiv1.Topic, hash string) (ctrl.Result, error) {
	topic.Labels = util.SetHashToLabel(topic.Labels, hash)
	err := r.Update(ctx, topic)
	if err != nil {
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *TopicReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&apiv1.Topic{}).
		Complete(r)
}
