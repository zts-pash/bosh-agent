package integration_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry/bosh-agent/integration/integrationagentclient"
	"github.com/cloudfoundry/bosh-agent/settings"
)

var _ = Describe("fetch_logs", func() {
	var (
		agentClient      *integrationagentclient.IntegrationAgentClient
		registrySettings settings.Settings
	)

	BeforeEach(func() {
		err := testEnvironment.StopAgent()
		Expect(err).ToNot(HaveOccurred())

		err = testEnvironment.CleanupDataDir()
		Expect(err).ToNot(HaveOccurred())

		err = testEnvironment.CleanupLogFile()
		Expect(err).ToNot(HaveOccurred())

		err = testEnvironment.SetupConfigDrive()
		Expect(err).ToNot(HaveOccurred())

		err = testEnvironment.UpdateAgentConfig("config-drive-agent.json")
		Expect(err).ToNot(HaveOccurred())

		registrySettings = settings.Settings{
			AgentID: "fake-agent-id",

			// note that this SETS the username and password for HTTP message bus access
			Mbus: "https://mbus-user:mbus-pass@127.0.0.1:6868",

			Env: settings.Env{
				Bosh: settings.BoshEnv{
					TargetedBlobstores: settings.TargetedBlobstores{
						Packages: "custom-blobstore",
						Logs:     "custom-blobstore",
					},
					Blobstores: []settings.Blobstore{
						settings.Blobstore{
							Type: "local",
							Name: "ignored-blobstore",
							Options: map[string]interface{}{
								"blobstore_path": "/ignored/blobstore",
							},
						},
						settings.Blobstore{
							Type: "local",
							Name: "custom-blobstore",
							Options: map[string]interface{}{
								// this path should get rewritten internally to /var/vcap/data/blobs
								"blobstore_path": "/var/vcap/micro_bosh/data/cache",
							},
						},
					},
				},
			},

			Disks: settings.Disks{
				Ephemeral: "/dev/sdh",
			},
		}

		err = testEnvironment.AttachDevice("/dev/sdh", 128, 2)
		Expect(err).ToNot(HaveOccurred())

		err = testEnvironment.StartRegistry(registrySettings)
		Expect(err).ToNot(HaveOccurred())

		err = testEnvironment.StartAgent()
		Expect(err).ToNot(HaveOccurred())

		agentClient, err = testEnvironment.StartAgentTunnel("mbus-user", "mbus-pass", 6868)
		Expect(err).NotTo(HaveOccurred())

		_, err = testEnvironment.RunCommand("sudo mkdir -p /var/vcap/data")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := testEnvironment.StopAgentTunnel()
		Expect(err).NotTo(HaveOccurred())

		err = testEnvironment.StopAgent()
		Expect(err).NotTo(HaveOccurred())

		err = testEnvironment.DetachDevice("/dev/sdh")
		Expect(err).ToNot(HaveOccurred())
	})

	It("puts the logs in the appropriate blobstore location", func() {
		r, stderr, _, err := testEnvironment.RunCommand3("echo 'foobarbaz' | sudo tee /var/vcap/sys/log/fetch-logs")
		Expect(err).NotTo(HaveOccurred(), r, stderr)

		logsResponse, err := agentClient.FetchLogs("job", nil)
		Expect(err).NotTo(HaveOccurred())

		r, stderr, _, err = testEnvironment.RunCommand3(fmt.Sprintf("sudo zcat /var/vcap/data/blobs/%s", logsResponse["blobstore_id"]))

		Expect(err).NotTo(HaveOccurred(), r, stderr)
		Expect(r).To(ContainSubstring("foobarbaz"))
		Expect(r).To(ContainSubstring("fetch-logs"))
	})
})
