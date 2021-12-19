package tunnel

import (
	corev1 "github.com/cokeos/zero/api/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	LabelKey   = "cokeos.io/zero-managed"
	LabelValue = "true"

	UniqLabelKey = "cokeos.io/zero-id"
)

func generateService(tunnel *corev1.Tunnel) *v1.Service {
	return &v1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      tunnel.GetName(),
			Namespace: tunnel.GetNamespace(),
			Labels: map[string]string{
				LabelKey: LabelValue,
			},
		},
		Spec: v1.ServiceSpec{
			Type: v1.ServiceTypeNodePort,
			Selector: map[string]string{
				UniqLabelKey: tunnel.Spec.UnitName,
			},
			Ports: tunnel.Spec.Ports,
		},
	}
}
