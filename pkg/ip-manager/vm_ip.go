package ip_manager

import (
	"fmt"
	kubevirt "kubevirt.io/api/core/v1"
)

func VmNamespaced(machine *kubevirt.VirtualMachine) string {
	return fmt.Sprintf("vm/%s/%s", machine.Namespace, machine.Name)
}
