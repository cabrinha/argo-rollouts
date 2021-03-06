package rollout

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	patchtypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/controller"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/argoproj/argo-rollouts/utils/annotations"
	"github.com/argoproj/argo-rollouts/utils/conditions"
	logutil "github.com/argoproj/argo-rollouts/utils/log"
	replicasetutil "github.com/argoproj/argo-rollouts/utils/replicaset"
)

const (
	switchSelectorPatch = `{
	"spec": {
		"selector": {
			"%s": "%s"
		}
	}
}`
)

// switchSelector switch the selector on an existing service to a new value
func (c RolloutController) switchServiceSelector(service *corev1.Service, newRolloutUniqueLabelValue string, r *v1alpha1.Rollout) error {
	if service.Spec.Selector == nil {
		service.Spec.Selector = make(map[string]string)
	}
	if oldPodHash, ok := service.Spec.Selector[v1alpha1.DefaultRolloutUniqueLabelKey]; ok && oldPodHash == newRolloutUniqueLabelValue {
		return nil
	}
	patch := fmt.Sprintf(switchSelectorPatch, v1alpha1.DefaultRolloutUniqueLabelKey, newRolloutUniqueLabelValue)
	_, err := c.kubeclientset.CoreV1().Services(service.Namespace).Patch(service.Name, patchtypes.StrategicMergePatchType, []byte(patch))
	if err != nil {
		return err
	}
	msg := fmt.Sprintf("Switched selector for service '%s' to value '%s'", service.Name, newRolloutUniqueLabelValue)
	logutil.WithRollout(r).Info(msg)
	c.recorder.Event(r, corev1.EventTypeNormal, "SwitchService", msg)
	service.Spec.Selector[v1alpha1.DefaultRolloutUniqueLabelKey] = newRolloutUniqueLabelValue
	return err
}

func (c *RolloutController) reconcilePreviewService(roCtx *blueGreenContext, previewSvc *corev1.Service) error {
	r := roCtx.Rollout()
	logCtx := roCtx.Log()
	newRS := roCtx.NewRS()
	if previewSvc == nil {
		return nil
	}
	logCtx.Infof("Reconciling preview service '%s'", previewSvc.Name)

	newPodHash := newRS.Labels[v1alpha1.DefaultRolloutUniqueLabelKey]
	err := c.switchServiceSelector(previewSvc, newPodHash, r)
	if err != nil {
		return err
	}

	return nil
}

func (c *RolloutController) reconcileActiveService(roCtx *blueGreenContext, previewSvc, activeSvc *corev1.Service) error {
	r := roCtx.Rollout()
	newRS := roCtx.NewRS()
	allRSs := roCtx.AllRSs()

	if !replicasetutil.ReadyForPause(r, newRS, allRSs) || !annotations.IsSaturated(r, newRS) {
		roCtx.log.Infof("New RS '%s' is not fully saturated", newRS.Name)
		return nil
	}

	newPodHash := activeSvc.Spec.Selector[v1alpha1.DefaultRolloutUniqueLabelKey]
	//
	if skipPause(roCtx, activeSvc) {
		newPodHash = newRS.Labels[v1alpha1.DefaultRolloutUniqueLabelKey]
	}
	if roCtx.PauseContext().CompletedBlueGreenPause() && completedPrePromotionAnalysis(roCtx) {
		newPodHash = newRS.Labels[v1alpha1.DefaultRolloutUniqueLabelKey]
	}

	if r.Status.Abort {
		currentRevision := int(0)
		for _, rs := range controller.FilterActiveReplicaSets(roCtx.OlderRSs()) {
			revision := replicasetutil.GetReplicaSetRevision(r, rs)
			if revision > currentRevision {
				newPodHash = rs.Labels[v1alpha1.DefaultRolloutUniqueLabelKey]
				currentRevision = revision
			}
		}
	}

	err := c.switchServiceSelector(activeSvc, newPodHash, r)
	if err != nil {
		return err
	}
	return nil
}

// getReferencedService returns service references in rollout spec and sets warning condition if service does not exist
func (c *RolloutController) getReferencedService(r *v1alpha1.Rollout, serviceName string) (*corev1.Service, error) {
	svc, err := c.servicesLister.Services(r.Namespace).Get(serviceName)
	if err != nil {
		if errors.IsNotFound(err) {
			msg := fmt.Sprintf(conditions.ServiceNotFoundMessage, serviceName)
			c.recorder.Event(r, corev1.EventTypeWarning, conditions.ServiceNotFoundReason, msg)
			newStatus := r.Status.DeepCopy()
			cond := conditions.NewRolloutCondition(v1alpha1.RolloutProgressing, corev1.ConditionFalse, conditions.ServiceNotFoundReason, msg)
			c.patchCondition(r, newStatus, cond)
		}
		return nil, err
	}
	return svc, nil
}

func (c *RolloutController) getPreviewAndActiveServices(r *v1alpha1.Rollout) (*corev1.Service, *corev1.Service, error) {
	var previewSvc *corev1.Service
	var activeSvc *corev1.Service
	var err error

	if r.Spec.Strategy.BlueGreen.PreviewService != "" {
		previewSvc, err = c.getReferencedService(r, r.Spec.Strategy.BlueGreen.PreviewService)
		if err != nil {
			return nil, nil, err
		}
	}
	if r.Spec.Strategy.BlueGreen.ActiveService == "" {
		return nil, nil, fmt.Errorf("Invalid Spec: Rollout missing field ActiveService")
	}
	activeSvc, err = c.getReferencedService(r, r.Spec.Strategy.BlueGreen.ActiveService)
	if err != nil {
		return nil, nil, err
	}
	return previewSvc, activeSvc, nil
}

func (c *RolloutController) reconcileStableAndCanaryService(roCtx *canaryContext) error {
	r := roCtx.Rollout()
	newRS := roCtx.NewRS()
	stableRS := roCtx.StableRS()
	if r.Spec.Strategy.Canary == nil {
		return nil
	}
	if r.Spec.Strategy.Canary.StableService != "" && stableRS != nil {
		svc, err := c.getReferencedService(r, r.Spec.Strategy.Canary.StableService)
		if err != nil {
			return err
		}
		if svc.Spec.Selector[v1alpha1.DefaultRolloutUniqueLabelKey] != stableRS.Labels[v1alpha1.DefaultRolloutUniqueLabelKey] {
			err = c.switchServiceSelector(svc, stableRS.Labels[v1alpha1.DefaultRolloutUniqueLabelKey], r)
			if err != nil {
				return err
			}
		}

	}
	if r.Spec.Strategy.Canary.CanaryService != "" && newRS != nil {
		svc, err := c.getReferencedService(r, r.Spec.Strategy.Canary.CanaryService)
		if err != nil {
			return err
		}
		if svc.Spec.Selector[v1alpha1.DefaultRolloutUniqueLabelKey] != newRS.Labels[v1alpha1.DefaultRolloutUniqueLabelKey] {
			err = c.switchServiceSelector(svc, newRS.Labels[v1alpha1.DefaultRolloutUniqueLabelKey], r)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
