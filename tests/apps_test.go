package tests

import (
	"strings"

	deis "github.com/deis/controller-sdk-go"
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/builds"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"
	"github.com/deis/workflow-e2e/tests/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis apps", func() {

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

			Context("and another user also exists", func() {

				var otherUser model.User

				BeforeEach(func() {
					otherUser = auth.Register()
				})

				AfterEach(func() {
					auth.Cancel(otherUser)
				})

				Specify("that first user can transfer ownership to the other user", func() {
					sess, err := cmd.Start("deis apps:transfer --app=%s %s", &user, app.Name, otherUser.Username)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
					sess, err = cmd.Start("deis info -a %s", &user, app.Name)
					Eventually(sess.Err).Should(Say(util.PrependError(deis.ErrForbidden)))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(1))
					// Transer back or else cleanup will fail.
					sess, err = cmd.Start("deis apps:transfer --app=%s %s", &otherUser, app.Name, user.Username)
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

			})

		})

		Context("who owns an existing app that has already been deployed", func() {

			uuidRegExp := `[0-9a-f]{8}-([0-9a-f]{4}-){3}[0-9a-f]{12}`
			procsRegexp := `(%s-[\w-]+) up \(v\d+\)`

			var app model.App

			BeforeEach(func() {
				app = apps.Create(user, "--no-remote")
				builds.Create(user, app)
			})

			AfterEach(func() {
				apps.Destroy(user, app)
			})

			Specify("that user can get information about that app", func() {
				sess, err := cmd.Start("deis info -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Application", app.Name))
				Eventually(sess).Should(Say(`uuid:\s*%s`, uuidRegExp))
				Eventually(sess).Should(Say(`url:\s*%s`, strings.Replace(app.URL, "http://", "", 1)))
				Eventually(sess).Should(Say(`owner:\s*%s`, user.Username))
				Eventually(sess).Should(Say(`id:\s*%s`, app.Name))
				Eventually(sess).Should(Say("=== %s Processes", app.Name))
				Eventually(sess).Should(Say(procsRegexp, app.Name))
				Eventually(sess).Should(Say("=== %s Domains", app.Name))
				Eventually(sess).Should(Say("%s", app.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can retrieve logs for that app", func() {
				sess, err := cmd.Start("deis logs -a %s", &user, app.Name)
				Eventually(sess).Should(SatisfyAll(
					Say("INFO \\[%s\\]: %s created initial release", app.Name, user.Username),
					Say("INFO \\[%s\\]: domain %s added", app.Name, app.Name)))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can open that app", func() {
				apps.Open(user, app)
			})

			Specify("that user can run a command in that app's environment", func() {
				sess, err := cmd.Start("deis apps:run --app=%s echo Hello, 世界", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess, (settings.MaxEventuallyTimeout)).Should(Say("Hello, 世界"))
				Eventually(sess).Should(Exit(0))
			})

		})

	})

})
