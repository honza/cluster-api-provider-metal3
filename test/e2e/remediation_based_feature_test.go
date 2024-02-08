package e2e

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

/*
 * Remediation-based Tests
 * This test focus on verifying the effectiveness of fixes or remedial actions taken to address node failures.
 * These tests involve simulating failure scenarios, triggering the remediation process, and then verifying that the remediation actions successfully restore the nodes to the desired state.
 *
 * Test Types:
 * 1. Metal3Remediation Test: This test specifically evaluates the Metal3 Remediation Controller's node deletion feature in the reboot remediation strategy.
 * 2. Remediation Test: This test focuses on verifying various annotations and actions related to remediation in the CAPM3 (Cluster API Provider for Metal3).
 *
 * NodeDeletionRemediation Test:
 * - Retrieve the list of Metal3 machines associated with the worker nodes.
 * - Identify the target worker Metal3Machine and its corresponding BareMetalHost (BMH) object.
 * - Create a Metal3Remediation resource with a remediation strategy of type "Reboot" and a specified timeout.
 * - Wait for the associated virtual machine (VM) to power off.
 * - Wait for the node (VM) to be deleted.
 * - Wait for the VM to power on.
 * - Wait for the node to be in a ready state.
 * - Delete the Metal3Remediation resource.
 * - Verify that the Metal3Remediation resource has been successfully deleted.
 *
 * Healthcheck Test:
 * - For both worker and controlplane machines:
 * - Create and deploy machinehealthcheck.
 * - Stop kubelet on the machine.
 * - Wait for the healthcheck to notice the unhealthy machine.
 * - Wait for the remediation request to be created.
 * - Wait for the machine to appear healthy again.
 * - Wait until the remediation request has been deleted.
 *
 * Remediation Test:
 * - Reboot Annotation: Mark a worker BMH for reboot and wait for the associated VM to transition to the "shutoff" state and then to the "running" state.
 * - Poweroff Annotation: Verify the power off and power on actions by turning off and on the specified machines.
 * - Inspection Annotation: Run an inspection test alongside the remediation steps to verify the inspection annotation functionality.
 * - Unhealthy Annotation: Mark a BMH as unhealthy and ensure it is not picked up for provisioning.
 * - Metal3 Data Template: Create a new Metal3DataTemplate (M3DT), create a new Metal3MachineTemplate (M3MT), and update the MachineDeployment (MD) to point to the new M3MT. Wait for the old worker to deprovision.
 */
var _ = Describe("Testing nodes remediation [remediation] [features]", func() {

	var (
		ctx                 = context.TODO()
		specName            = "metal3"
		namespace           = "metal3"
		clusterName         = "test1"
		clusterctlLogFolder string
	)

	BeforeEach(func() {
		osType := strings.ToLower(os.Getenv("OS"))
		Expect(osType).ToNot(Equal(""))
		validateGlobals(specName)

		// We need to override clusterctl apply log folder to avoid getting our credentials exposed.
		clusterctlLogFolder = filepath.Join(os.TempDir(), "clusters", bootstrapClusterProxy.GetName())
	})

	It("Should create a cluster and run remediation based tests", func() {
		By("Creating target cluster")
		targetCluster, _ = createTargetCluster(e2eConfig.GetVariable("KUBERNETES_VERSION"))

		// Run Metal3Remediation test first, doesn't work after remediation...
		By("Running node deletion remediation tests")
		nodeDeletionRemediation(ctx, func() NodeDeletionRemediation {
			return NodeDeletionRemediation{
				E2EConfig:             e2eConfig,
				BootstrapClusterProxy: bootstrapClusterProxy,
				TargetCluster:         targetCluster,
				SpecName:              specName,
				ClusterName:           clusterName,
				Namespace:             namespace,
			}
		})

		By("Running healthcheck tests")
		healthcheck(ctx, func() HealthCheckInput {
			return HealthCheckInput{
				BootstrapClusterProxy: bootstrapClusterProxy,
				ClusterName:           clusterName,
				Namespace:             namespace,
			}
		})

		By("Running annotated powercycle remediation tests")
		remediation(ctx, func() RemediationInput {
			return RemediationInput{
				E2EConfig:             e2eConfig,
				BootstrapClusterProxy: bootstrapClusterProxy,
				TargetCluster:         targetCluster,
				SpecName:              specName,
				ClusterName:           clusterName,
				Namespace:             namespace,
				ClusterctlConfigPath:  clusterctlConfigPath,
			}
		})
	})

	AfterEach(func() {
		ListBareMetalHosts(ctx, bootstrapClusterProxy.GetClient(), client.InNamespace(namespace))
		ListMetal3Machines(ctx, bootstrapClusterProxy.GetClient(), client.InNamespace(namespace))
		ListMachines(ctx, bootstrapClusterProxy.GetClient(), client.InNamespace(namespace))
		ListNodes(ctx, targetCluster.GetClient())
		// Abort the test in case of failure and keepTestEnv is true during keep VM trigger
		if CurrentSpecReport().Failed() {
			if keepTestEnv {
				AbortSuite("e2e test aborted and skip cleaning the VM", 4)
			}
		}
		DumpSpecResourcesAndCleanup(ctx, specName, bootstrapClusterProxy, artifactFolder, namespace, e2eConfig.GetIntervals, clusterName, clusterctlLogFolder, skipCleanup)
	})

})
