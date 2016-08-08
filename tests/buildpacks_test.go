package tests

import (
	"fmt"
	"os"
	"strings"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/git"
	"github.com/deis/workflow-e2e/tests/cmd/keys"
	"github.com/deis/workflow-e2e/tests/model"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("all buildpack apps", func() {

	Context("with an existing user", func() {

		var user model.User
		var keyPath string

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Context("who has added their public key", func() {

			BeforeEach(func() {
				_, keyPath = keys.Add(user)
			})

			DescribeTable("can deploy an example buildpack app",
				func(url, buildpack, banner string) {

					var app model.App

					output, err := cmd.Execute(`git clone %s`, url)
					Expect(err).NotTo(HaveOccurred(), output)
					// infer app directory from URL
					splits := strings.Split(url, "/")
					dir := strings.TrimSuffix(splits[len(splits)-1], ".git")
					os.Chdir(dir)
					// create with custom buildpack if needed
					var args []string
					if buildpack != "" {
						args = append(args, fmt.Sprintf("--buildpack %s", buildpack))
					}
					app = apps.Create(user, args...)
					defer apps.Destroy(user, app)
					git.Push(user, keyPath, app, banner)

				},

				// NOTE: Keep this list up-to-date with any example apps that are added
				// under the github/deis org, or any third-party apps that increase coverage
				// or prevent regressions.
				Entry("Go", "https://github.com/deis/example-go.git", "",
					"Powered by Deis"),
			)

		})

	})

})
