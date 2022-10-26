package ephemeral

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func VolumeClaimName(pod *v1.Pod, volume *v1.Volume) string {
	return pod.Name + "-" + volume.Name
}

func VolumeIsForPod(pod *v1.Pod, pvc *v1.PersistentVolumeClaim) error {
	// Checking the namespaces is just a precaution. The caller should
	// never pass in a PVC that isn't from the same namespace as the
	// Pod.
	if pvc.Namespace != pod.Namespace || !metav1.IsControlledBy(pvc, pod) {
		return fmt.Errorf("PVC %s/%s was not created for pod %s/%s (pod is not owner)", pvc.Namespace, pvc.Name, pod.Namespace, pod.Name)
	}
	return nil
}
