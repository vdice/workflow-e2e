package tests

import (
	deis "github.com/deis/controller-sdk-go"
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"
	"github.com/deis/workflow-e2e/tests/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis perms", func() {

	Context("with an existing admin", func() {

		admin := model.Admin

		Specify("that admin can list admins", func() {
			sess, err := cmd.Start("deis perms:list --admin", &admin)
			Eventually(sess).Should(Say("=== Administrators"))
			Eventually(sess).Should(Say(admin.Username))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(0))
		})

		Context("and another existing user", func() {

			var otherUser model.User

			BeforeEach(func() {
				otherUser = auth.Register()
			})

			AfterEach(func() {
				auth.Cancel(otherUser)
			})

			Specify("that admin can grant admin permissions to the other user", func() {
				sess, err := cmd.Start("deis perms:create %s --admin", &admin, otherUser.Username)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Adding %s to system administrators... done\n", otherUser.Username))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis perms:list --admin", &admin)
				Eventually(sess).Should(Say("=== Administrators"))
				Eventually(sess).Should(Say(otherUser.Username))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Context("who owns an existing app", func() {

				var app model.App

				BeforeEach(func() {
					app = apps.Create(otherUser, "--no-remote")
				})

				AfterEach(func() {
					apps.Destroy(otherUser, app)
				})

				Specify("that admin can list permissions on the app owned by the second user", func() {
					sess, err := cmd.Start("deis perms:list --app=%s", &admin, app.Name)
					Eventually(sess).Should(Say("=== %s's Users", app.Name))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

			})

		})

	})

	Context("with an existing non-admin user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Specify("that user cannot list admin permissions", func() {
			sess, err := cmd.Start("deis perms:list --admin", &user)
			Eventually(sess.Err).Should(Say(util.PrependError(deis.ErrForbidden)))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(1))
		})

		Specify("that user cannot create admin permissions", func() {
			sess, err := cmd.Start("deis perms:create %s --admin", &user, user.Username)
			Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("Adding %s to system administrators...", user.Username))
			Eventually(sess.Err).Should(Say(util.PrependError(deis.ErrForbidden)))
			Expect(err).NotTo(HaveOccurred())
			Eventually(sess).Should(Exit(1))
		})

		Context("and an existing admin", func() {

			admin := model.Admin

			Specify("the non-admin user cannot delete the admin's admin permissions", func() {
				sess, err := cmd.Start("deis perms:delete %s --admin", &user, admin.Username)
				Eventually(sess.Err, settings.MaxEventuallyTimeout).Should(Say(util.PrependError(deis.ErrForbidden)))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

		})

	})

})
