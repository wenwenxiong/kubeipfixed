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

package virtualmachine

import (
	"context"
	ip_manager "github.com/wenwenxiong/kubeipfixed/pkg/ip-manager"
	"gomodules.xyz/jsonpatch/v2"
	admissionv1 "k8s.io/api/admission/v1"
	kubevirt "kubevirt.io/api/core/v1"
	"math/rand"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	kawwebhook "github.com/qinqon/kube-admission-webhook/pkg/webhook"
)

var log = logf.Log.WithName("Webhook mutatevirtualmachines")

type virtualMachineAnnotator struct {
	client      client.Client
	decoder     *admission.Decoder
	poolManager *ip_manager.IPManager
}

// Add adds server modifiers to the server, like registering the hook to the webhook server.
func Add(s *kawwebhook.Server, poolManager *ip_manager.IPManager) error {
	virtualMachineAnnotator := &virtualMachineAnnotator{poolManager: poolManager}
	s.Register("/mutate-virtualmachines", &webhook.Admission{Handler: virtualMachineAnnotator})
	return nil
}

// podAnnotator adds an annotation to every incoming pods.
func (a *virtualMachineAnnotator) Handle(ctx context.Context, req admission.Request) admission.Response {
	virtualMachine := &kubevirt.VirtualMachine{}

	err := a.decoder.Decode(req, virtualMachine)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	handleRequestId := rand.Intn(100000)
	logger := log.WithName("Handle").WithValues("RequestId", handleRequestId, "virtualMachineFullName", ip_manager.VmNamespaced(virtualMachine))

	if virtualMachine.Annotations == nil {
		virtualMachine.Annotations = map[string]string{}
	}
	if virtualMachine.Namespace == "" {
		virtualMachine.Namespace = req.AdmissionRequest.Namespace
	}

	logger.V(1).Info("got a virtual machine event")

	// admission.PatchResponse generates a Response containing patches.
	kubemapcoolJsonPatches := []jsonpatch.Operation{}
	return admission.Response{
		Patches: kubemapcoolJsonPatches,
		AdmissionResponse: admissionv1.AdmissionResponse{
			Allowed:   true,
			PatchType: func() *admissionv1.PatchType { pt := admissionv1.PatchTypeJSONPatch; return &pt }(),
		},
	}
}
