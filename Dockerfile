FROM ubuntu-debootstrap:14.04

COPY tests/tests.test .
RUN mv tests.test /bin
RUN apt-get update -y && apt-get install -y curl openssh-client git
RUN curl -sSL http://deis.io/deis-cli/install-v2-alpha.sh | bash && mv ./deis /bin/deis
CMD /bin/tests.test
