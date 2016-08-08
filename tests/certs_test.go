package tests

import (
	"fmt"
	"net/http"
	"strconv"

	deis "github.com/deis/controller-sdk-go"
	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/builds"
	"github.com/deis/workflow-e2e/tests/cmd/certs"
	"github.com/deis/workflow-e2e/tests/cmd/domains"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/util"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis certs", func() {

	var cert model.Cert

	BeforeEach(func() {
		cert = model.NewCert()
	})

	Context("with an existing user", func() {

		var user model.User

		BeforeEach(func() {
			user = auth.Register()
		})

		AfterEach(func() {
			auth.Cancel(user)
		})

		Context("who owns an existing cert", func() {

			nonExistentDomain := "non.existent.domain"

			BeforeEach(func() {
				certs.Add(user, cert)
			})

			AfterEach(func() {
				certs.Remove(user, cert)
			})

			Specify("that user cannot attach a cert to a non-existent domain", func() {
				sess, err := cmd.Start("deis certs:attach %s %s", &user, cert.Name, nonExistentDomain)
				Eventually(sess.Err).Should(Say(util.PrependError(deis.ErrNotFound)))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

			Specify("that user cannot detach a cert from a non-existent domain", func() {
				sess, err := cmd.Start("deis certs:detach %s %s", &user, cert.Name, nonExistentDomain)
				Eventually(sess.Err).Should(Say(util.PrependError(deis.ErrNotFound)))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(1))
			})

		})

		Context("who owns two existing certs", func() {

			var cert1, cert2 model.Cert

			BeforeEach(func() {
				cert1 = model.NewCert()
				cert2 = model.NewCert()
				certs.Add(user, cert1)
				certs.Add(user, cert2)
			})

			AfterEach(func() {
				certs.Remove(user, cert1)
				certs.Remove(user, cert2)
			})

			Specify("that user can limit the number of certs returned by certs:list", func() {
				randCertRegExp := `\d{0,9}-cert`

				// limit=0 is invalid as of DRF 3.4
				// https://github.com/tomchristie/django-rest-framework/pull/4194
				sess, err := cmd.Start("deis certs:list --limit=0", &user)
				Eventually(sess).Should(Say(randCertRegExp))
				Eventually(sess).Should(Say(randCertRegExp))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis certs:list --limit=1", &user)
				Eventually(sess).Should(Say(randCertRegExp))
				Eventually(sess).Should(Not(Say(randCertRegExp)))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))

				sess, err = cmd.Start("deis certs:list", &user)
				Eventually(sess).Should(Say(randCertRegExp))
				Eventually(sess).Should(Say(randCertRegExp))
				Expect(err).NotTo(HaveOccurred())
				Eventually(sess).Should(Exit(0))
			})

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

			Context("with a domain added to it", func() {

				domain := "www.foo.com"

				BeforeEach(func() {
					domains.Add(user, app, domain)
				})

				AfterEach(func() {
					domains.Remove(user, app, domain)
				})

				Context("and that user also owns an existing cert", func() {

					BeforeEach(func() {
						certs.Add(user, cert)
					})

					AfterEach(func() {
						certs.Remove(user, cert)
					})

					Specify("that user can attach/detach that cert to/from that domain", func() {
						certs.Attach(user, cert, domain)
						curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -k -H "Host: %s" -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, domain, app.URL)}
						Eventually(cmd.Retry(curlCmd, strconv.Itoa(http.StatusOK), 60)).Should(BeTrue())
						certs.Detach(user, cert, domain)
					})

				})

			})

		})

	})

})
