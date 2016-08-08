package tests

import (
	"regexp"
	"strings"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/builds"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis tags", func() {

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Context("who owns an existing app that has already been deployed", func() {

			var app model.App

			BeforeEach(func() {
				app = apps.Create(user, "--no-remote")
				builds.Create(user, app)
			})

			AfterEach(func() {
				apps.Destroy(user, app)
			})

			Context("and a tag has already been added to the app", func() {

				var label []string

				BeforeEach(func() {
					// Find a valid tag to set
					// Use original $HOME dir or else kubectl can't find its config
					sess, err := cmd.Start("HOME=%s kubectl get nodes -o jsonpath={.items[*].metadata..labels}", nil, settings.ActualHome)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					// grep output like "map[kubernetes.io/hostname:192.168.64.2 node:worker1]"
					re := regexp.MustCompile(`([\w\.\-]{0,253}/?[-_\.\w]{1,63}:[-_\.\w]{1,63})`)
					pairs := re.FindAllString(string(sess.Out.Contents()), -1)
					// Use the first key:value pair found
					label = strings.Split(pairs[0], ":")

					sess, err = cmd.Start("deis tags:set --app=%s %s=%s", &user, app.Name, label[0], label[1])
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Tags", app.Name))
					Eventually(sess).Should(Say(`%s\s+%s`, label[0], label[1]))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

				Specify("that user can unset that tag from that app", func() {
					sess, err := cmd.Start("deis tags:unset --app=%s %s", &user, app.Name, label[0])
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Tags", app.Name))
					Eventually(sess).ShouldNot(Say(`%s\s+%s`, label[0], label[1]))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					sess, err = cmd.Start("deis tags:list --app=%s", &user, app.Name)
					Eventually(sess).Should(Say("=== %s Tags", app.Name))
					Eventually(sess).ShouldNot(Say(`%s\s+%s`, label[0], label[1]))
					Eventually(sess).ShouldNot(Say(`munkafolyamat\s+yeah`, app.Name))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

			})

		})

	})

})
