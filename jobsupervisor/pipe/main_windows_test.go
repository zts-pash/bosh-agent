package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Main", func() {
	Context("when parent exits", func() {
		It("kills children and exits", func(done Done) {
			cmd := exec.Command(ExitRunnerPath, pathToPipeCLI, PrintPidsPath)
			cmd.Env = append(os.Environ(),
				joinEnv("SERVICE_NAME", ServiceName),
			)
			var stdout bytes.Buffer
			cmd.Stdout = &stdout

			Expect(cmd.Start()).To(Succeed())
			Eventually(func() string { return stdout.String() }).Should(ContainSubstring(","))
			pids := strings.Split(strings.TrimSpace(stdout.String()), ",")

			i, err := strconv.Atoi(pids[0])
			Expect(err).To(Succeed())
			pipeProc, err := os.FindProcess(i)
			Expect(err).To(Succeed())

			i, err = strconv.Atoi(pids[1])
			Expect(err).To(Succeed())
			childProc, err := os.FindProcess(i)
			Expect(err).To(Succeed())

			st, err := pipeProc.Wait()
			fmt.Println(st)
			Expect(err).To(Succeed())

			_, err = childProc.Wait()
			Expect(err).To(Succeed())

			close(done)
		}, 10)
	})
})
