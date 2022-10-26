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

package ip_manager

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"

	netattdefv1 "github.com/k8snetworkplumbingwg/network-attachment-definition-client/pkg/apis/k8s.cni.cncf.io/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	kubevirt "kubevirt.io/api/core/v1"
)

const tempPodName = "tempPodName"

func (p *IPManager) AllocatePodIP(pod *corev1.Pod, isNotDryRun bool) error {
	p.poolMutex.Lock()
	defer p.poolMutex.Unlock()

	networkValue, ok := pod.Annotations[sriovNetworksAnnotation]
	if !ok {
		return nil
	}

	networks, err := parsePodNetworkAnnotation(networkValue, pod.Namespace)
	if err != nil {
		return err
	}

	log.V(1).Info("pod meta data", "podMetaData", (*pod).ObjectMeta)

	if len(networks.ipPool) == 0 {
		return nil
	}

	// validate if the pod is related to kubevirt
	if p.isRelatedToKubevirt(pod) {
		// nothing to do here. the mac is already by allocated by the virtual machine webhook
		log.V(1).Info("This pod have ownerReferences from kubevirt skipping")
		return nil
	}

	for _, network := range networks.ipPool {
		raw, err := network.RenderNetAttDef(networks.resourceName)
		if err != nil {
			return err
		}
		netAttDef := &netattdefv1.NetworkAttachmentDefinition{}

		err = p.Scheme.Convert(raw, netAttDef, nil)
		if err != nil {
			return err
		}
		// Check if this NetworkAttachmentDefinition already exists
		found := &netattdefv1.NetworkAttachmentDefinition{}
		err = p.kubeClient.Get(context.TODO(), types.NamespacedName{Name: netAttDef.Name, Namespace: netAttDef.Namespace}, found)
		if err != nil {
			if errors.IsNotFound(err) {
				log.V(1).Info("NetworkAttachmentDefinition CR not exist, creating")
				err = p.kubeClient.Create(context.TODO(), netAttDef)
				if err != nil {
					log.V(1).Error(err, "Couldn't create NetworkAttachmentDefinition CR", "Namespace", netAttDef.Namespace, "Name", netAttDef.Name)
					return err
				}
			} else {
				log.V(1).Error(err, "Couldn't get NetworkAttachmentDefinition CR", "Namespace", netAttDef.Namespace, "Name", netAttDef.Name)
				return err
			}
		} else {
			log.V(1).Info("NetworkAttachmentDefinition CR already exist")
			if !reflect.DeepEqual(found.Spec, netAttDef.Spec) || !reflect.DeepEqual(found.GetAnnotations(), netAttDef.GetAnnotations()) {
				log.V(1).Info("Update NetworkAttachmentDefinition CR", "Namespace", netAttDef.Namespace, "Name", netAttDef.Name)
				netAttDef.SetResourceVersion(found.GetResourceVersion())
				err = p.kubeClient.Update(context.TODO(), netAttDef)
				if err != nil {
					log.V(1).Error(err, "Couldn't update NetworkAttachmentDefinition CR", "Namespace", netAttDef.Namespace, "Name", netAttDef.Name)
					return err
				}
			}
		}
	}

	name := networks.ipPool[0].Name
	namespace := networks.ipPool[0].Namespace

	networkListJson := "[{\"name\": \"" + name + "\", \"namespace\":\"" + namespace + "\"}]"
	pod.Annotations[NetworksAnnotation] = networkListJson

	return nil
}

func (p *IPManager) isRelatedToKubevirt(pod *corev1.Pod) bool {
	if pod.ObjectMeta.OwnerReferences == nil {
		return false
	}

	for _, ref := range pod.OwnerReferences {
		if ref.Kind == kubevirt.VirtualMachineInstanceGroupVersionKind.Kind {
			vm := &kubevirt.VirtualMachine{}
			err := p.kubeClient.Get(context.TODO(), client.ObjectKey{Namespace: pod.Namespace, Name: ref.Name}, vm)
			if err != nil && apierrors.IsNotFound(err) {
				log.V(1).Info("this pod is an ephemeral vmi object allocating mac as a regular pod")
				return false
			}

			return true
		}
	}

	return false
}

func parsePodNetworkAnnotation(podNetworks, defaultNamespace string) (*sriovNetwork, error) {
	var networks *sriovNetwork

	if podNetworks == "" {
		return nil, fmt.Errorf("parsePodNetworkAnnotation: pod annotation not having \"sriovNetworks\" as key, refer  README.md for the usage guide")
	}

	if strings.IndexAny(podNetworks, "[{\"") >= 0 {
		if err := json.Unmarshal([]byte(podNetworks), &networks); err != nil {
			return nil, fmt.Errorf("parsePodNetworkAnnotation: failed to parse pod Network Attachment Selection Annotation JSON format: %v", err)
		}
	} else {
		log.Info("Only JSON List Format for networks is allowed to be parsed")
		return networks, nil
	}

	for _, sriovIp := range networks.ipPool {
		if sriovIp.Namespace == "" {
			sriovIp.Namespace = defaultNamespace
		}
		if sriovIp.Nameservers == "" {
			sriovIp.Nameservers = defaultNameservers
		}
	}

	return networks, nil
}
