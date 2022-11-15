/*
Copyright 2022.

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

package tenancykcpdev

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	tenancyv1alpha1 "github.com/kcp-dev/kcp/pkg/apis/tenancy/v1alpha1"
	"github.com/kcp-dev/logicalcluster/v2"
	sharedutil "github.com/redhat-appstudio/managed-gitops/backend-shared/util"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

// ClusterWorkspaceReconciler reconciles a ClusterWorkspace object
type ClusterWorkspaceReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=tenancy.kcp.dev,resources=clusterworkspaces,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=tenancy.kcp.dev,resources=clusterworkspaces/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=tenancy.kcp.dev,resources=clusterworkspaces/finalizers,verbs=update
//+kubebuilder:rbac:groups=apis.kcp.dev,resources=apiexports,verbs=get;list;watch
//+kubebuilder:rbac:groups=apis.kcp.dev,resources=apiexports/content,verbs=*

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ClusterWorkspaceType object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.0/pkg/reconcile
func (r *ClusterWorkspaceReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := r.Log.WithValues("ClusterWorkspace", req.NamespacedName).WithValues("clusterName", req.ClusterName)

	// If the request has no cluster name, it isn't a KCP workspace, so exit
	if req.ClusterName == "" {
		return ctrl.Result{}, nil
	}

	// Get the ClusterWorkspace
	clusterName := fmt.Sprintf("%s:%s", req.ClusterName, req.Name)
	logger.Info("the expected clusterName", "clusterName", clusterName)

	ctx = logicalcluster.WithCluster(ctx, logicalcluster.New(clusterName))

	// Create the ServiceAccount
	serviceAccountName := "TODO"
	serviceAccountNS := "default"
	sa, err := sharedutil.GetOrCreateServiceAccount(ctx, r.Client, serviceAccountName, serviceAccountNS, logger)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to create service account %s: %w", serviceAccountName, err)
	}

	// Create the ServiceAccount Secret
	tokenSecret, err := sharedutil.CreateServiceAccountTokenSecret(ctx, r.Client, serviceAccountName, serviceAccountNS, logger)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to create service account bearer token: %w", err)
	}
	// Add annotation to secret identifying that it was created by the CWT
	patch := client.MergeFrom(tokenSecret.DeepCopy())
	tokenSecret.ObjectMeta.Annotations["TODO"] = "todo"
	r.Client.Patch(ctx, tokenSecret, patch)

	logger.Info("created sa %s:", sa.Name) // TODO: remove this

	// Create ClusterRole and ClusterRoleBinding
	err = sharedutil.CreateOrUpdateClusterRoleAndRoleBinding(ctx, "TODO-UID", r.Client, serviceAccountName, serviceAccountNS, logger)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("unable to create cluster role and rolebinding: %w", err)
	}

	return ctrl.Result{}, nil
}

/*func (r *ClusterWorkspaceReconciler) createServiceAccount(ctx context.Context, name string) (*corev1.ServiceAccount, error) {
	serviceAccount := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
			//Annotations:

		},
		Secrets: []corev1.ObjectReference{
			{
				Name: "TODO",
			},
		},
	}
	err := r.Client.Create(ctx, serviceAccount)
	if err != nil {
		return nil, err
	}
	return serviceAccount, nil
}*/

func gitopsEnvironmentCreatedPredicate() predicate.Predicate {
	return predicate.Funcs{
		CreateFunc: func(e event.CreateEvent) bool {
			if e.Object.GetAnnotations()["appstudio.redhat.com/clusterworkspacetype"] == "gitops-environment" {
				return true
			}
			return false
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return false
		},
		GenericFunc: func(e event.GenericEvent) bool {
			return false
		},
		UpdateFunc: func(e event.UpdateEvent) bool {
			return false
		},
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterWorkspaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&tenancyv1alpha1.ClusterWorkspace{}).
		//WithEventFilter(gitopsEnvironmentCreatedPredicate()).
		Complete(r)
}
