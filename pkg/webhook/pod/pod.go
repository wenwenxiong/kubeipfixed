/*
Copyright 2019 The KubeMacPool Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package pod

import (
	"context"
	ip_manager "github.com/wenwenxiong/kubeipfixed/pkg/ip-manager"
	"gomodules.xyz/jsonpatch/v2"
	"net/http"

	admissionv1 "k8s.io/api/admission/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	kawwebhook "github.com/qinqon/kube-admission-webhook/pkg/webhook"
)

var log = logf.Log.WithName("Webhook mutatepods")

type podAnnotator struct {
	client    client.Client
	decoder   *admission.Decoder
	ipManager *ip_manager.IPManager
}

// Add adds server modifiers to the server, like registering the hook to the webhook server.
func Add(s *kawwebhook.Server, ipManager *ip_manager.IPManager) error {
	podAnnotator := &podAnnotator{ipManager: ipManager}
	s.Register("/mutate-pods", &webhook.Admission{Handler: podAnnotator})
	return nil
}

// Handle podAnnotator adds an annotation to every incoming pods.
func (a *podAnnotator) Handle(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	err := a.decoder.Decode(req, pod)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}
	originalPod := pod.DeepCopy()

	if pod.Annotations == nil {
		pod.Annotations = map[string]string{}
	}

	isNotDryRun := (req.DryRun == nil || *req.DryRun == false)
	transactionTimestamp := ip_manager.CreateTransactionTimestamp()
	log.V(1).Info("got a create pod event", "podName", pod.Name, "podNamespace", pod.Namespace, "transactionTimestamp", transactionTimestamp)

	err = a.ipManager.AllocatePodIP(pod, isNotDryRun)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	// admission.PatchResponse generates a Response containing patches.
	return patchPodChanges(originalPod, pod)
}

// create jsonpatches only to changed caused by the kubeipfixed webhook changes
func patchPodChanges(originalPod, currentPod *corev1.Pod) admission.Response {
	kubeIPJsonPatches := []jsonpatch.Operation{}

	currentNetworkAnnotation := currentPod.GetAnnotations()[ip_manager.NetworksAnnotation]
	originalPodNetworkAnnotation := originalPod.GetAnnotations()[ip_manager.NetworksAnnotation]
	if originalPodNetworkAnnotation != currentNetworkAnnotation {
		annotationPatch := jsonpatch.NewOperation("replace", "/metadata/annotations", currentPod.GetAnnotations())
		kubeIPJsonPatches = append(kubeIPJsonPatches, annotationPatch)
	}

	log.Info("patchPodChanges", "kubemapcoolJsonPatches", kubeIPJsonPatches)
	if len(kubeIPJsonPatches) == 0 {
		return admission.Response{
			Patches: kubeIPJsonPatches,
			AdmissionResponse: admissionv1.AdmissionResponse{
				Allowed: true,
			},
		}
	}
	return admission.Response{
		Patches: kubeIPJsonPatches,
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed:   true,
			PatchType: func() *admissionv1.PatchType { pt := admissionv1.PatchTypeJSONPatch; return &pt }(),
		},
	}
}

// InjectClient injects the client into the podAnnotator
func (a *podAnnotator) InjectClient(c client.Client) error {
	a.client = c
	return nil
}

// InjectDecoder injects the decoder.
func (a *podAnnotator) InjectDecoder(d *admission.Decoder) error {
	a.decoder = d
	return nil
}
