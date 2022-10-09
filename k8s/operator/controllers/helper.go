package controllers

import (
	tunnelv1 "github.com/sebgrz/operator/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type tunnelAgentDeploymentConfig struct {
	Name        string
	Namespace   string
	Destination string
	Resource    *tunnelv1.TunnelAgent
}

func deploymentName(name string) string {
	return name + "-deployment"
}

func createTunnelAgentDeployment(config tunnelAgentDeploymentConfig) *appsv1.Deployment {
	var replicas int32 = 1
	deployment := &appsv1.Deployment{
		ObjectMeta: v1.ObjectMeta{
			Name:      deploymentName(config.Name),
			Namespace: config.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &v1.LabelSelector{
				MatchLabels: map[string]string{"app": config.Name},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: v1.ObjectMeta{
					Labels: map[string]string{"app": config.Name},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    config.Name,
							Image:   "ghcr.io/sebgrz/tunnel:master", // TODO: take from TunnelAgent CRD
							Command: []string{"/app/cmd/agent"},
							Args: []string{
								"-server", config.Resource.Spec.Server,
								"-hostname", config.Resource.Spec.Hostname,
								"-destination", config.Destination,
								"-type", config.Resource.Spec.Type,
							},
						},
					},
				},
			},
		},
	}
	return deployment
}
