package datadog

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/argoproj/argo-rollouts/utils/evaluate"
	metricutil "github.com/argoproj/argo-rollouts/utils/metric"
	templateutil "github.com/argoproj/argo-rollouts/utils/template"
	log "github.com/sirupsen/logrus"
	dd "github.com/zorkian/go-datadog-api"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const (
	// ProviderType indicates the provider is datadog
	ProviderType = "Datadog"
	// DatadogSecretName is a k8s secret that holds datadog api and app keys
	DatadogSecretName = "datadog-api-keys"
)

// Provider contains all the required components to run a datadog query
type Provider struct {
	client dd.Client
	logCtx log.Entry
}

// Type incidates provider is a datadog provider
func (p *Provider) Type() string {
	return ProviderType
}

// Run queries DataDog for a metric
func (p *Provider) Run(run *v1alpha1.AnalysisRun, metric v1alpha1.Metric) v1alpha1.Measurement {
	startTime := metav1.Now()
	newMeasurement := v1alpha1.Measurement{
		StartedAt: &startTime,
	}

	query, err := templateutil.ResolveArgs(metric.Provider.Datadog.Query, run.Spec.Args)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	// TODO (cabrinha) make from and to configurable
	from := time.Now().Unix() - 60
	to := time.Now().Unix()
	response, err := p.client.QueryMetrics(from, to, query)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	// If datadog returns an empty []Series, retry with a longer window
	var retries int
	for retries = 0; retries < 3 && len(response) < 1; retries++ {
		response, err = p.client.QueryMetrics(from-60, to, query)
		if err != nil {
			return metricutil.MarkMeasurementError(newMeasurement, err)
		}

		if len(response) < 1 {
			time.Sleep(1 * time.Second)
		}
	}

	if len(response) > 0 {
		newValue, newStatus, err := p.processResponse(metric, response[0])
		if err != nil {
			return metricutil.MarkMeasurementError(newMeasurement, err)
		}
		newMeasurement.Value = newValue
		newMeasurement.Phase = newStatus
		finishedTime := metav1.Now()
		newMeasurement.FinishedAt = &finishedTime
		return newMeasurement
	}
	newMeasurement.Value = ""
	newMeasurement.Phase = v1alpha1.AnalysisPhaseInconclusive
	finishedTime := metav1.Now()
	newMeasurement.FinishedAt = &finishedTime
	return newMeasurement
}

// Resume should not be used with the datadog provider since all the work should occur in the Run method
func (p *Provider) Resume(run *v1alpha1.AnalysisRun, metric v1alpha1.Metric, measurement v1alpha1.Measurement) v1alpha1.Measurement {
	p.logCtx.Warn("Datadog provider should not execute the Resume method")
	return measurement
}

// Terminate should not be used with the datadog provider since all the work should occur in the Run method
func (p *Provider) Terminate(run *v1alpha1.AnalysisRun, metric v1alpha1.Metric, measurement v1alpha1.Measurement) v1alpha1.Measurement {
	p.logCtx.Warn("Datadog provider should not execute the Terminate method")
	return measurement
}

// GarbageCollect is a no-op for the datadog provider
func (p *Provider) GarbageCollect(run *v1alpha1.AnalysisRun, metric v1alpha1.Metric, limit int) error {
	return nil
}

func (p *Provider) processResponse(metric v1alpha1.Metric, response dd.Series) (string, v1alpha1.AnalysisPhase, error) {
	length := len(response.Points)
	switch {
	case length < 1:
		return "", v1alpha1.AnalysisPhaseInconclusive, nil
	// If we get 1 or more points, select the first point
	case length >= 1:
		result := *response.Points[0][1]
		valueStr := fmt.Sprintf("%f", result)
		newStatus := evaluate.EvaluateResult(result, metric, p.logCtx)
		return valueStr, newStatus, nil
	// TODO (cabrinha) add other response types
	default:
		return "", v1alpha1.AnalysisPhaseError, fmt.Errorf("Datadog metric type not supported")
	}
}

// NewDatadogProvider creates a new Datadog client
func NewDatadogProvider(client dd.Client, logCtx log.Entry) *Provider {
	return &Provider{
		logCtx: logCtx,
		client: client,
	}
}

// NewDatadogAPI generates a datadog API from the metric configuration
func NewDatadogAPI(metric v1alpha1.Metric, kubeclientset kubernetes.Interface) (*dd.Client, error) {
	ns := Namespace()
	secret, err := kubeclientset.CoreV1().Secrets(ns).Get(DatadogSecretName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if len(secret.Data[`datadog-api-key`]) > 0 && len(secret.Data[`datadog-app-key`]) > 0 {
		apiKey := fmt.Sprintf("%s", secret.Data[`datadog-api-key`])
		appKey := fmt.Sprintf("%s", secret.Data[`datadog-app-key`])
		client := dd.NewClient(apiKey, appKey)

		if metric.Provider.Datadog.BaseURL != "" {
			client.SetBaseUrl(metric.Provider.Datadog.BaseURL)
		}

		_, err := client.Validate()
		if err != nil {
			return nil, err
		}
		return client, nil
	}
	return nil, errors.New("failed to make client: no datadog API or App keys found")
}

// Namespace will return $POD_NAMESPACE if set, otherwise "argo-rollouts"
func Namespace() string {
	// This way assumes you've set the POD_NAMESPACE environment variable using the downward API.
	// This check has to be done first for backwards compatibility with the way InClusterConfig was originally set up
	if ns, ok := os.LookupEnv("POD_NAMESPACE"); ok {
		return ns
	}
	// Fall back to the namespace associated with the service account token, if available
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns
		}
	}
	return "argo-rollouts"
}
