package datadog

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/argoproj/argo-rollouts/pkg/apis/rollouts/v1alpha1"
	"github.com/argoproj/argo-rollouts/utils/controller"
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
	measurement := p.runMeasurement(run, metric)
	return measurement
}

// Resume executes Run if the analysis phase is inconclusive or until 3 retries have been executed
func (p *Provider) Resume(run *v1alpha1.AnalysisRun, metric v1alpha1.Metric, measurement v1alpha1.Measurement) v1alpha1.Measurement {
	measurement = p.runMeasurement(run, metric)
	var retries int
	for retries = 0; retries < 3 && measurement.Phase != v1alpha1.AnalysisPhaseSuccessful; retries++ {
		measurement = p.runMeasurement(run, metric)
		if measurement.Value == "" {
			time.Sleep(1 * time.Second)
		}
		return measurement
	}
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

func (p *Provider) runMeasurement(run *v1alpha1.AnalysisRun, metric v1alpha1.Metric) v1alpha1.Measurement {
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
	newValue, newStatus, err := p.processResponse(metric, response)
	if err != nil {
		return metricutil.MarkMeasurementError(newMeasurement, err)
	}

	if newValue == "" && newStatus != v1alpha1.AnalysisPhaseSuccessful {
		resumeTime := metav1.NewTime(time.Now().Add(2 * time.Second))
		newMeasurement.FinishedAt = nil
		newMeasurement.Phase = v1alpha1.AnalysisPhaseRunning
		newMeasurement.ResumeAt = &resumeTime
	} else if newValue != "" && newStatus == v1alpha1.AnalysisPhaseSuccessful {
		finishedTime := metav1.Now()
		newMeasurement.FinishedAt = &finishedTime
		newMeasurement.Phase = newStatus
	}
	newMeasurement.Value = newValue
	return newMeasurement
}

// flattenResponse removes the unix timestamp from a []DataPoint, flattens all values into a []float64
// and returns a []float64 and a string of [float64, ...]
func flattenResponse(dp []dd.DataPoint) ([]float64, string) {
	floats := make([]float64, len(dp))

	if len(dp) > 2 {
		valueStr := "["
		for _, v := range dp {
			floats = append(floats, *v[1])
			valueStr = valueStr + fmt.Sprintf("%.2f", *v[1]) + ","
		}

		if len(valueStr) > 1 {
			valueStr = valueStr[:len(valueStr)-1] + "]" // strip last comma
		} else {
			valueStr = ""
		}
		return floats, valueStr
	}

	valueStr := fmt.Sprintf("%.2f", *dp[0][1])
	return floats, valueStr
}

func (p *Provider) processResponse(metric v1alpha1.Metric, response []dd.Series) (string, v1alpha1.AnalysisPhase, error) {
	if len(response) == 0 {
		return "", v1alpha1.AnalysisPhaseInconclusive, nil
	} else if len(response) >= 1 {
		series := response[0]
		results, valueStr := flattenResponse(series.Points)
		for _, result := range results {
			if math.IsNaN(result) {
				return valueStr, v1alpha1.AnalysisPhaseInconclusive, nil
			}
		}
		newStatus := evaluate.EvaluateResult(results, metric, p.logCtx)
		return valueStr, newStatus, nil
	}

	return "", v1alpha1.AnalysisPhaseFailed, fmt.Errorf("No data points found in response from Datadog")
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
	ns := controller.Namespace()
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