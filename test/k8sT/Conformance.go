// SPDX-License-Identifier: Apache-2.0
// Copyright 2020-2021 Authors of Cilium

package k8sTest

import (
	. "github.com/onsi/gomega"

	. "github.com/cilium/cilium/test/ginkgo-ext"
	"github.com/cilium/cilium/test/helpers"
)

var _ = Describe("K8sConformance", func() {
	SkipContextIf(func() bool {
		return helpers.RunsWithKubeProxyReplacement()
	}, "Portmap Chaining", func() {
		var (
			kubectl                         *helpers.Kubectl
			ciliumFilename                  string
			connectivityCheckYaml           string
			connectivityCheckYamlQuarantine string
			connectivityCheckYamlSimple     string
		)

		BeforeAll(func() {
			kubectl = helpers.CreateKubectl(helpers.K8s1VMName(), logger)
			connectivityCheckYaml = kubectl.GetFilePath("../examples/kubernetes/connectivity-check/connectivity-check-hostport.yaml")
			connectivityCheckYamlQuarantine = kubectl.GetFilePath("../examples/kubernetes/connectivity-check/connectivity-check-quarantine.yaml")
			connectivityCheckYamlSimple = kubectl.GetFilePath("../examples/kubernetes/connectivity-check/connectivity-check-single-node.yaml")

			deployOpts := map[string]string{
				"cni.chainingMode": "portmap",
				// When kube-proxy is enabled, the host firewall is not
				// compatible with portmap chaining because traffic
				// from pods to remote nodes goes through the tunnel.
				// This issue is tracked at #12541.
				"hostFirewall.enabled": "false",
			}
			ciliumFilename = helpers.TimestampFilename("cilium.yaml")
			DeployCiliumOptionsAndDNS(kubectl, ciliumFilename, deployOpts)

			_, err := kubectl.CiliumNodesWait()
			ExpectWithOffset(1, err).Should(BeNil(), "Failure while waiting for k8s nodes to be annotated by Cilium")

			By("Making sure all endpoints are in ready state")
			err = kubectl.CiliumEndpointWaitReady()
			ExpectWithOffset(1, err).To(BeNil(), "Failure while waiting for all cilium endpoints to reach ready state")
		})

		AfterEach(func() {
			kubectl.Delete(connectivityCheckYaml)
			kubectl.Delete(connectivityCheckYamlQuarantine)
			kubectl.Delete(connectivityCheckYamlSimple)
			ExpectAllPodsInNsTerminated(kubectl, "default")
		})

		AfterFailed(func() {
			kubectl.CiliumReport("cilium endpoint list")
		})

		AfterAll(func() {
			UninstallCiliumFromManifest(kubectl, ciliumFilename)
			kubectl.CloseSSHClient()
		})

		JustAfterEach(func() {
			kubectl.ValidateNoErrorsInLogs(CurrentGinkgoTestDescription().Duration)
		})

		It("Check connectivity-check compliance with portmap chaining", func() {
			kubectl.ApplyDefault(connectivityCheckYaml).ExpectSuccess("cannot install connectivity-check")

			err := kubectl.WaitforPods(helpers.DefaultNamespace, "", helpers.HelperTimeout)
			ExpectWithOffset(1, err).Should(BeNil(), "connectivity-check pods are not ready after timeout")
		})

		It("Check one node connectivity-check compliance with portmap chaining", func() {
			kubectl.ApplyDefault(connectivityCheckYamlSimple).ExpectSuccess("cannot install connectivity-check-single-node")

			err := kubectl.WaitforPods(helpers.DefaultNamespace, "", helpers.HelperTimeout)
			ExpectWithOffset(1, err).Should(BeNil(), "connectivity-check pods are not ready after timeout")
		})

		// FIXME: GH-12700 L7 policy breaks connectivity to hostport services.
		//        When this is resolved, remove 'quarantine: true' label in examples/kubernetes/connectivity-check/proxy.cue.
		//        The tests will be merged into the above checks, so this test can be removed.
		XIt("Check connectivity-check compliance with proxy and portmap chaining", func() {
			kubectl.ApplyDefault(connectivityCheckYamlQuarantine).ExpectSuccess("cannot install connectivity-check")

			err := kubectl.WaitforPods(helpers.DefaultNamespace, "", helpers.HelperTimeout)
			ExpectWithOffset(1, err).Should(BeNil(), "connectivity-check pods are not ready after timeout")
		})
	})
})
