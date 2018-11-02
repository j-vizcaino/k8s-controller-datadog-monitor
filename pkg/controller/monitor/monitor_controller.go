package monitor

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/j-vizcaino/k8s-controller-datadog-monitor/pkg/datadog-client"
	log2 "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"time"

	apiv1alpha1 "github.com/j-vizcaino/k8s-controller-datadog-monitor/pkg/apis/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Monitor Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileMonitor{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		log: log2.ZapLogger(true).WithName("monitor-controller"),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	mgr.GetScheme().AddTypeDefaultingFunc(nil, nil)
	c, err := controller.New("monitor-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Monitor
	return c.Watch(&source.Kind{Type: &apiv1alpha1.Monitor{}}, &handler.EnqueueRequestForObject{})
}

var _ reconcile.Reconciler = &ReconcileMonitor{}

// ReconcileMonitor reconciles a Monitor object
type ReconcileMonitor struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	log    logr.Logger
}

const (
	ReconcileLoopPeriod = 5 * time.Second
	FinalizerName       = "api.datadoghq.com"
)

var nextMonitorID = 1

// Reconcile reads that state of the cluster for a Monitor object and makes changes based on the state read
// and what is in the Monitor.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileMonitor) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	res := reconcile.Result{}
	log := r.log.WithValues("name", fmt.Sprintf("%s/%s", request.Namespace, request.Name))

	// Fetch the Monitor monitor
	monitor := &apiv1alpha1.Monitor{}
	err := r.client.Get(context.TODO(), request.NamespacedName, monitor)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Resource not found")
			return res, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "resource get failed")
		return res, err
	}

	switch {

	// Monitor is already deleted, will be garbage collected later
	case monitor.Status.State == apiv1alpha1.MonitorStateDeleted:
		log.Info("Monitor is already deleted","id", monitor.Status.MonitorID)
		return res, nil

	// Monitor needs to be deleted
	case monitor.DeletionTimestamp != nil:
		log.Info("Deleting monitor", "id", monitor.Status.MonitorID)
		res, err = r.deleteMonitor(monitor, log)

	// Monitor needs to be checked/updated
	case monitor.Status.State == apiv1alpha1.MonitorStateCreated:
		res, err = r.updateMonitor(monitor, log)

	// New monitor
	default:
		res, err = r.createMonitor(monitor, log)
	}

	return res, err
}

func (r *ReconcileMonitor) updateMonitor(monitor *apiv1alpha1.Monitor, log logr.Logger) (reconcile.Result, error) {
	log = log.WithValues("id", monitor.Status.MonitorID)
	newChecksum := monitor.Spec.Checksum()
	if monitor.Status.LastAppliedChecksum == newChecksum {
		log.Info("Monitor content did not change, not updating")
		return reconcile.Result{}, nil
	}

	m := monitor.DeepCopy()
	m.Status.LastAppliedChecksum = newChecksum
	err := r.client.Update(context.TODO(), m)
	if err != nil {
		log.Error(err, "Failed to update monitor object")
	} else {
		log.Info("Monitor updated", "checksum", newChecksum)
	}
	return reconcile.Result{}, err
}

func (r *ReconcileMonitor) createMonitor(monitor *apiv1alpha1.Monitor, log logr.Logger) (reconcile.Result, error) {
	m := monitor.DeepCopy()
	c, ok := datadog_client.Registry[monitor.Spec.TargetSite]
	if !ok {
		m.Status = apiv1alpha1.MonitorStatus{
			State: apiv1alpha1.MonitorStateError,
			ErrorMessage: fmt.Sprintf("site '%s' not supported by controller", monitor.Spec.TargetSite),
		}
	} else {
		res, err := c.Post(context.TODO(), "/api/v1/monitor", monitor.Spec.Data)
		if err != nil {
			m.Status = apiv1alpha1.MonitorStatus{
				State: apiv1alpha1.MonitorStateError,
				ErrorMessage: err.Error(),
			}
		} else {
			m.Finalizers = append(m.Finalizers, FinalizerName)
			m.Status = apiv1alpha1.MonitorStatus{
				State: apiv1alpha1.MonitorStateCreated,
				MonitorID: int(res["id"].(float64)),
				LastAppliedChecksum: monitor.Spec.Checksum(),
			}
		}
	}

	err := r.client.Update(context.TODO(), m)
	if err == nil {
		log.Info("Monitor status updated", "id", m.Status.MonitorID)
	}
	return reconcile.Result{}, err
}

func (r *ReconcileMonitor) deleteMonitor(monitor *apiv1alpha1.Monitor, log logr.Logger) (reconcile.Result, error) {
	log = log.WithValues("id", monitor.Status.MonitorID)
	m := monitor.DeepCopy()

	found := false
	for idx, f := range monitor.Finalizers {
		if f == FinalizerName {
			found = true
			m.Finalizers = append(m.Finalizers[:idx], m.Finalizers[idx+1:]...)
			break
		}
	}

	if !found {
		log.Info("Missing identifier from finalizers","finalizers", monitor.Finalizers)
		return reconcile.Result{}, nil
	}
	log.Info("Removed identifier from finalizers")

	m.Status.MonitorID = 0
	m.Status.State = apiv1alpha1.MonitorStateDeleted
	err := r.client.Update(context.TODO(), m)
	return reconcile.Result{}, err
}
