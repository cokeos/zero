package unit

import (
	corev1 "github.com/cokeos/zero/api/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
	"strconv"
)

const (
	PythonEnvKey   = "PYTHONUNBUFFERED"
	PythonEnvValue = "0"

	SSH     = "ssh"
	SSHPort = 22

	DefaultMountPath   = "/data"
	DefaultGlusterPath = "/data"

	LabelKey   = "cokeos.io/zero-managed"
	LabelValue = "true"

	UniqLabelKey = "cokeos.io/zero-id"

	NodeGPUModelKey = "cokeos.io/gpu-model"

	DefaultGPUNumber = "0"

	DefaultGPUModel = "GTX-1660"
)

func defaultGpuModelList() []string {
	return []string{
		DefaultGPUModel,
	}
}

func getPodImage(unit *corev1.Unit) string {
	if unit.Spec.GPUPolicy.GPU {
		return "ccr.ccs.tencentyun.com/njupt-isl/" +
			unit.Spec.Framework.Name +
			"-gpu:" +
			unit.Spec.Framework.Version
	}
	return "ccr.ccs.tencentyun.com/njupt-isl/" +
		unit.Spec.Framework.Name +
		"-cpu:" +
		unit.Spec.Framework.Version
}

func generatePod(unit *corev1.Unit) *v1.Pod {
	// 环境变量检测
	env := unit.Spec.Execution.Env
	if len(env) == 0 {
		env = make([]v1.EnvVar, 0)
	}
	env = append(env, v1.EnvVar{
		Name:  PythonEnvKey,
		Value: PythonEnvValue,
	})
	// 亲和标签
	model := make([]string, 0)
	if unit.Spec.GPUPolicy.GPU {

		if unit.Spec.GPUPolicy.Model == "" {
			model = defaultGpuModelList()
		} else {
			model = append(model, unit.Spec.GPUPolicy.Model)
		}
	}
	affinity := &v1.Affinity{
		NodeAffinity: &v1.NodeAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: &v1.NodeSelector{
				NodeSelectorTerms: []v1.NodeSelectorTerm{
					{
						MatchExpressions: []v1.NodeSelectorRequirement{
							{
								Key:      NodeGPUModelKey,
								Operator: v1.NodeSelectorOpIn,
								Values:   model,
							},
						},
					},
				},
			},
		},
	}

	// gpu 检测
	gpu, err := resource.ParseQuantity(strconv.Itoa(unit.Spec.GPUPolicy.Number))
	if err != nil {
		klog.Error(err)
	}
	if !unit.Spec.GPUPolicy.GPU {
		gpu, err = resource.ParseQuantity(DefaultGPUNumber)
		if err != nil {
			klog.Error(err)
		}
	}

	// 端口检测
	ports := make([]v1.ContainerPort, 0)
	if len(unit.Spec.Ports) == 0 {
		ports = append(ports, v1.ContainerPort{
			Name:          SSH,
			ContainerPort: SSHPort,
		})
	}

	// 执行命令检测
	command := unit.Spec.Execution.Command
	args := unit.Spec.Execution.Args
	if unit.Spec.Execution.SSH {
		command = nil
		args = nil
	}

	return &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: unit.Namespace,
			Name:      unit.Name,
			Labels: map[string]string{
				LabelKey:     LabelValue,
				UniqLabelKey: unit.Namespace + "/" + unit.Name,
			},
		},
		Spec: v1.PodSpec{
			Affinity:      affinity,
			RestartPolicy: v1.RestartPolicyNever,
			Containers: []v1.Container{
				{
					Name:    unit.Name,
					Image:   getPodImage(unit),
					Env:     env,
					Ports:   ports,
					Command: command,
					Args:    args,
					Resources: v1.ResourceRequirements{
						Limits: map[v1.ResourceName]resource.Quantity{
							v1.ResourceCPU:           unit.Spec.ResourceList.Cpu().DeepCopy(),
							v1.ResourceMemory:        unit.Spec.ResourceList.Memory().DeepCopy(),
							corev1.ResourceNvidiaGPU: gpu,
						},
					},
					VolumeMounts: []v1.VolumeMount{
						{
							Name:      unit.Name + "-vol",
							MountPath: DefaultMountPath,
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: unit.Name + "-vol",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: DefaultGlusterPath + unit.Namespace,
						},
					},
				},
			},
		},
	}
}
