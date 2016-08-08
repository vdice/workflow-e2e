package tests

import (
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

var _ = Describe("deis healthchecks", func() {

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

			Specify("that user can list healthchecks on that app", func() {
				sess, err := cmd.Start("deis healthchecks:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Healthchecks", app.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set an exec liveness healthcheck", func() {
				sess, err := cmd.Start("deis healthchecks:set -a %s liveness exec -- /bin/true", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Applying livenessProbe healthcheck..."))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`Exec Probe\: Command=\[/bin/true]`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis healthchecks:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`Exec Probe\: Command=\[/bin/true]`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set an exec readiness healthcheck", func() {
				sess, err := cmd.Start("deis healthchecks:set readiness exec -a %s -- /bin/true", &user, app.Name)
				Eventually(sess).Should(Say("Applying readinessProbe healthcheck..."))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`Exec Probe\: Command=\[/bin/true]`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis healthchecks:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Healthchecks", app.Name))
				Eventually(sess).Should(Say(`Exec Probe\: Command=\[/bin/true]`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})
		})
	})
})
