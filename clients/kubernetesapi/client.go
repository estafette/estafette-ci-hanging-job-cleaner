package kubernetesapi

import (
	"context"

	"github.com/opentracing/opentracing-go"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client interface {
	GetJobs(ctx context.Context) (jobs []batchv1.Job, err error)
	GetConfigMaps(ctx context.Context) (configmaps []v1.ConfigMap, err error)
	GetSecrets(ctx context.Context) (secrets []v1.Secret, err error)

	DeleteJob(ctx context.Context, job batchv1.Job) (err error)
	DeleteConfigMap(ctx context.Context, configmap v1.ConfigMap) (err error)
	DeleteSecret(ctx context.Context, secret v1.Secret) (err error)
}

// NewClient returns a new kubernetesapi.Client
func NewClient(namespace string) (Client, error) {

	// create kubernetes api client
	kubeClientConfig, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	// creates the clientset
	kubeClientset, err := kubernetes.NewForConfig(kubeClientConfig)
	if err != nil {
		return nil, err
	}

	return &client{
		kubeClientset: kubeClientset,
		namespace:     namespace,
	}, nil
}

type client struct {
	kubeClientset *kubernetes.Clientset
	namespace     string
}

func (c *client) GetJobs(ctx context.Context) (jobs []batchv1.Job, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "kubernetesapi.Client:GetJobs")
	defer span.Finish()

	log.Info().Msgf("Retrieving jobs with label createdBy=estafette in namespace %v...", c.namespace)

	jobsList, err := c.kubeClientset.BatchV1().Jobs(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "createdBy=estafette",
	})
	if err != nil {
		return
	}

	jobs = jobsList.Items

	log.Info().Msgf("Retrieved %v jobs with label createdBy=estafette in namespace %v", len(jobs), c.namespace)

	return jobs, nil
}

func (c *client) GetConfigMaps(ctx context.Context) (configmaps []v1.ConfigMap, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "kubernetesapi.Client:GetConfigMaps")
	defer span.Finish()

	log.Info().Msgf("Retrieving configmaps with label createdBy=estafette in namespace %v...", c.namespace)

	configmapsList, err := c.kubeClientset.CoreV1().ConfigMaps(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "createdBy=estafette",
	})
	if err != nil {
		return
	}

	configmaps = configmapsList.Items

	log.Info().Msgf("Retrieved %v configmaps with label createdBy=estafette in namespace %v", len(configmaps), c.namespace)

	return configmaps, nil
}

func (c *client) GetSecrets(ctx context.Context) (secrets []v1.Secret, err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "kubernetesapi.Client:GetSecrets")
	defer span.Finish()

	log.Info().Msgf("Retrieving secrets with label createdBy=estafette in namespace %v...", c.namespace)

	secretsList, err := c.kubeClientset.CoreV1().Secrets(c.namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "createdBy=estafette",
	})
	if err != nil {
		return
	}

	secrets = secretsList.Items

	log.Info().Msgf("Retrieved %v secrets with label createdBy=estafette in namespace %v", len(secrets), c.namespace)

	return secrets, nil
}

func (c *client) DeleteJob(ctx context.Context, job batchv1.Job) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "kubernetesapi.Client:DeleteSecret")
	defer span.Finish()

	log.Info().Msgf("Deleting job %v in namespace %v started at %v...", job.Name, c.namespace, job.CreationTimestamp.Time)

	propagationPolicy := metav1.DeletePropagationForeground
	err = c.kubeClientset.BatchV1().Jobs(c.namespace).Delete(ctx, job.Name, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		return
	}

	return nil
}

func (c *client) DeleteConfigMap(ctx context.Context, configmap v1.ConfigMap) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "kubernetesapi.Client:DeleteConfigMap")
	defer span.Finish()

	log.Info().Msgf("Deleting configmap %v in namespace %v started at %v...", configmap.Name, c.namespace, configmap.CreationTimestamp.Time)

	propagationPolicy := metav1.DeletePropagationForeground
	err = c.kubeClientset.CoreV1().ConfigMaps(c.namespace).Delete(ctx, configmap.Name, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		return
	}

	return nil
}

func (c *client) DeleteSecret(ctx context.Context, secret v1.Secret) (err error) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "kubernetesapi.Client:DeleteSecret")
	defer span.Finish()

	log.Info().Msgf("Deleting secret %v in namespace %v started at %v...", secret.Name, c.namespace, secret.CreationTimestamp.Time)

	propagationPolicy := metav1.DeletePropagationForeground
	err = c.kubeClientset.CoreV1().Secrets(c.namespace).Delete(ctx, secret.Name, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
	if err != nil {
		return
	}

	return nil
}
