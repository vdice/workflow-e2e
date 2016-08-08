package tests

import (
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis registry", func() {

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Context("who owns an existing app", func() {

			var app model.App

			BeforeEach(func() {
				app = apps.Create(user, "--no-remote")
			})

			AfterEach(func() {
				apps.Destroy(user, app)
			})

			Specify("that user can set a valid registry information", func() {
				// Setting a port first is required for a private registry
				sess, err := cmd.Start("deis config:set -a %s PORT=5000", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Creating config"))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
				Eventually(sess).Should(Say(`PORT\s+5000`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis registry:set --app=%s username=bob", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Registry", app.Name))
				Eventually(sess).Should(Say(`username\s+bob`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis registry:list --app=%s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Registry", app.Name))
				Eventually(sess).Should(Say(`username\s+bob`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Context("and registry information has already been added to the app", func() {

				BeforeEach(func() {
					// Setting a port first is required
					sess, err := cmd.Start("deis config:set -a %s PORT=5000", &user, app.Name)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Say("Creating config"))
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
					Eventually(sess).Should(Say(`PORT\s+5000`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					sess, err = cmd.Start("deis registry:set --app=%s username=bob", &user, app.Name)
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Registry", app.Name))
					Eventually(sess).Should(Say(`username\s+bob`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

				Specify("that user can unset that registry information from that app", func() {
					sess, err := cmd.Start("deis registry:unset --app=%s username", &user, app.Name)
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Registry", app.Name))
					Eventually(sess).ShouldNot(Say(`username\s+bob`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					sess, err = cmd.Start("deis registry:list --app=%s", &user, app.Name)
					Eventually(sess).Should(Say("=== %s Registry", app.Name))
					Eventually(sess).ShouldNot(Say(`username\s+bob`))
					Eventually(sess).ShouldNot(Say(`munkafolyamat\s+yeah`, app.Name))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

			})

		})

	})

	DescribeTable("any user can get command-line help for registry", func(command string, expected string) {
		sess, err := cmd.Start(command, nil)
		Eventually(sess).Should(Say(expected))
		Expect(err).NotTo(HaveOccurred())
		Eventually(sess).Should(Exit(0))
		// TODO: test that help output was more than five lines long
	},
		Entry("helps on \"help registry\"",
			"deis help registry", "Valid commands for registry:"),
		Entry("helps on \"registry -h\"",
			"deis registry -h", "Valid commands for registry:"),
		Entry("helps on \"registry --help\"",
			"deis registry --help", "Valid commands for registry:"),
		Entry("helps on \"help registry:list\"",
			"deis help registry:list", "Lists registry information for an application."),
		Entry("helps on \"registry:list -h\"",
			"deis registry:list -h", "Lists registry information for an application."),
		Entry("helps on \"registry:list --help\"",
			"deis registry:list --help", "Lists registry information for an application."),
		Entry("helps on \"help registry:set\"",
			"deis help registry:set", "Sets registry information for an application."),
		Entry("helps on \"registry:set -h\"",
			"deis registry:set -h", "Sets registry information for an application."),
		Entry("helps on \"registry:set --help\"",
			"deis registry:set --help", "Sets registry information for an application."),
		Entry("helps on \"help registry:unset\"",
			"deis help registry:unset", "Unsets registry information for an application."),
		Entry("helps on \"registry:unset -h\"",
			"deis registry:unset -h", "Unsets registry information for an application."),
		Entry("helps on \"registry:unset --help\"",
			"deis registry:unset --help", "Unsets registry information for an application."),
	)

})
