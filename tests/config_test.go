package tests

import (
	"io/ioutil"

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

var _ = Describe("deis config", func() {

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

			Specify("that user can list environment variables on that app", func() {
				sess, err := cmd.Start("deis config:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Config", app.Name))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set environment variables on that app", func() {
				sess, err := cmd.Start("deis config:set -a %s POWERED_BY=midi-chlorians", &user, app.Name)
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Say("Creating config"))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
				Eventually(sess).Should(Say(`POWERED_BY\s+midi-chlorians`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis config:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Config", app.Name))
				Eventually(sess).Should(Say(`POWERED_BY\s+midi-chlorians`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis run env -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("POWERED_BY=midi-chlorians"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Specify("that user can set an environment variable with non-ASCII and multibyte chars on that app", func() {
				sess, err := cmd.Start("deis config:set FOO=讲台 BAR=Þorbjörnsson BAZ=ноль -a %s", &user, app.Name)
				Eventually(sess).Should(Say("Creating config"))
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
				output := string(sess.Out.Contents())
				Expect(output).To(MatchRegexp(`FOO\s+讲台`))
				Expect(output).To(MatchRegexp(`BAR\s+Þorbjörnsson`))
				Expect(output).To(MatchRegexp(`BAZ\s+ноль`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis config:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Config", app.Name))
				output = string(sess.Out.Contents())
				Expect(output).To(MatchRegexp(`FOO\s+讲台`))
				Expect(output).To(MatchRegexp(`BAR\s+Þorbjörnsson`))
				Expect(output).To(MatchRegexp(`BAZ\s+ноль`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis run -a %s env", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
				output = string(sess.Out.Contents())
				Expect(output).To(ContainSubstring("FOO=讲台"))
				Expect(output).To(ContainSubstring("BAR=Þorbjörnsson"))
				Expect(output).To(ContainSubstring("BAZ=ноль"))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

			Context("and has already has an environment variable set", func() {

				BeforeEach(func() {
					sess, err := cmd.Start(`deis config:set -a %s FOO=xyzzy`, &user, app.Name)
					Eventually(sess).Should(Say("Creating config"))
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
					Eventually(sess).Should(Say(`FOO\s+xyzzy`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

				Specify("that user can unset that environment variable", func() {
					sess, err := cmd.Start("deis config:unset -a %s FOO", &user, app.Name)
					Eventually(sess).Should(Say("Removing config"))
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("=== %s Config", app.Name))
					Eventually(sess).ShouldNot(Say(`FOO\s+xyzzy`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					sess, err = cmd.Start("deis config:list -a %s", &user, app.Name)
					Eventually(sess).Should(Say("=== %s Config", app.Name))
					Eventually(sess).ShouldNot(Say(`FOO\s+xyzzy`))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					sess, err = cmd.Start("deis run -a %s env", &user, app.Name)
					Eventually(sess, settings.MaxEventuallyTimeout).ShouldNot(Say("FOO=xyzzy"))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

				Specify("that user can pull the configuration to an .env file", func() {
					sess, err := cmd.Start("deis config:pull -a %s", &user, app.Name)
					// TODO: ginkgo seems to redirect deis' file output here, so just examine
					// the output stream rather than reading in the .env file. Bug?
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say("FOO=xyzzy"))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))
				})

			})

			Specify("that user can push configuration from an .env file", func() {
				contents := []byte(`BIP=baz
FOO=bar`)
				err := ioutil.WriteFile(".env", contents, 0644)

				sess, err := cmd.Start("deis config:push -a %s", &user, app.Name)
				Eventually(sess, settings.MaxEventuallyTimeout).Should(Exit(0))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis config:list -a %s", &user, app.Name)
				Eventually(sess).Should(Say("=== %s Config", app.Name))
				Eventually(sess).Should(Say(`BIP\s+baz`))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

		})

	})

})
