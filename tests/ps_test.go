package tests

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"

	"github.com/deis/workflow-e2e/tests/cmd"
	"github.com/deis/workflow-e2e/tests/cmd/apps"
	"github.com/deis/workflow-e2e/tests/cmd/auth"
	"github.com/deis/workflow-e2e/tests/cmd/builds"
	"github.com/deis/workflow-e2e/tests/model"
	"github.com/deis/workflow-e2e/tests/settings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("deis ps", func() {

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

			DescribeTable("that user can scale that app up and down",
				func(scaleTo, respCode int) {
					sess, err := cmd.Start("deis ps:scale cmd=%d --app=%s", &user, scaleTo, app.Name)
					Eventually(sess).Should(Say("Scaling processes... but first,"))
					Eventually(sess, settings.MaxEventuallyTimeout).Should(Say(`done in \d+s`))
					Eventually(sess).Should(Say("=== %s Processes", app.Name))
					Expect(err).NotTo(HaveOccurred())
					Eventually(sess).Should(Exit(0))

					// test that there are the right number of processes listed
					procsListing := listProcs(user, app, "").Out.Contents()
					procs := scrapeProcs(app.Name, procsListing)
					Expect(len(procs)).To(Equal(scaleTo))

					// curl the app's root URL and print just the HTTP response code
					cmdRetryTimeout := 60
					curlCmd := model.Cmd{CommandLineString: fmt.Sprintf(`curl -sL -w "%%{http_code}\\n" "%s" -o /dev/null`, app.URL)}
					Eventually(cmd.Retry(curlCmd, strconv.Itoa(respCode), cmdRetryTimeout)).Should(BeTrue())
				},
				Entry("scales to 1", 1, 200),
				Entry("scales to 3", 3, 200),
				Entry("scales to 0", 0, 503),
			)

		})

	})

})

func listProcs(user model.User, app model.App, proctype string) *Session {
	sess, err := cmd.Start("deis ps:list --app=%s", &user, app.Name)
	Eventually(sess).Should(Say("=== %s Processes", app.Name))
	if proctype != "" {
		Eventually(sess).Should(Say("--- %s:", proctype))
	}
	Expect(err).NotTo(HaveOccurred())
	Eventually(sess).Should(Exit(0))
	return sess
}

// scrapeProcs returns the sorted process names for an app from the given output.
// It matches the current "deis ps" output for a healthy container:
//   earthy-vocalist-cmd-123456789-1d73e up (v2)
//   myapp-web-123456789-bujlq up (v16)
func scrapeProcs(app string, output []byte) []string {
	procsRegexp := `(%s-[\w-]+) up \(v\d+\)`
	re := regexp.MustCompile(fmt.Sprintf(procsRegexp, app))
	found := re.FindAllSubmatch(output, -1)
	procs := make([]string, len(found))
	for i := range found {
		procs[i] = string(found[i][1])
	}
	sort.Strings(procs)
	return procs
}
