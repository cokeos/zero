package tiny

import (
	corev1 "github.com/cokeos/zero/api/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func generateUnit(tiny *corev1.Tiny) *corev1.Unit {
	return &corev1.Unit{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tiny.GetNamespace(),
			Name:      tiny.GetName(),
		},
		Spec: corev1.UnitSpec{
			Framework: tiny.Spec.Framework,
			GPUPolicy: corev1.GPUPolicy{
				GPU:    tiny.Spec.GPU,
				Number: 1,
			},
			ResourceList: map[v1.ResourceName]resource.Quantity{
				v1.ResourceCPU:           resource.MustParse("1"),
				v1.ResourceMemory:        resource.MustParse("2Gi"),
				corev1.ResourceNvidiaGPU: resource.MustParse("1"),
			},
			Execution: corev1.Execution{
				SSH: true,
			},
		},
	}
}

const (
	SSH     = "ssh"
	SSHPort = 22
)

func generateTunnel(tiny *corev1.Tiny, port int32) *corev1.Tunnel {
	return &corev1.Tunnel{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: tiny.GetNamespace(),
			Name:      tiny.GetName(),
		},
		Spec: corev1.TunnelSpec{
			UnitName: tiny.GetNamespace() + "." + tiny.GetName(),
			Ports: []v1.ServicePort{
				{
					Name:       SSH,
					Protocol:   v1.ProtocolTCP,
					Port:       SSHPort,
					TargetPort: intstr.FromInt(SSHPort),
					NodePort:   port,
				},
			},
		},
	}
}
