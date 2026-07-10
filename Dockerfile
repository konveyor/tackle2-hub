ARG SEED_ROOT=/opt/app-root/src/tackle2-seed
ARG BRANDING=internal/frontend/auth/content/branding

FROM quay.io/konveyor/static-report AS report

FROM quay.io/centos/centos:stream9 AS centos
RUN dnf -y install epel-release && dnf -y install tini

FROM registry.access.redhat.com/ubi10/nodejs-22:latest AS frontend
ARG BRANDING
COPY --chown=1001:0 internal/frontend/auth/content/ .
COPY --chown=1001:0 ${BRANDING}/ branding/
RUN npm ci && npm run build

FROM registry.access.redhat.com/ubi10/go-toolset:latest AS builder
ENV GOPATH=$APP_ROOT
COPY --chown=1001:0 . .
COPY --chown=1001:0 --from=frontend /opt/app-root/src/dist/ internal/frontend/auth/content/dist/
RUN make docker
ARG SEED_PROJECT=konveyor/tackle2-seed
ARG SEED_BRANCH=main
ARG SEED_ROOT
RUN if [ ! -d "${SEED_ROOT}" ]; then \
      git clone --branch ${SEED_BRANCH} https://github.com/${SEED_PROJECT} ${SEED_ROOT}; \
    fi

FROM registry.access.redhat.com/ubi10/ubi-minimal
RUN mkdir -p /hub && chmod 0777 /hub
ENV HOME=/hub
WORKDIR /hub
ARG SEED_ROOT
ARG VERSION=latest
COPY --from=centos /usr/bin/tini /usr/bin/tini
COPY --from=builder /opt/app-root/src/bin/hub /usr/local/bin/tackle-hub
COPY --from=builder ${SEED_ROOT}/resources/ /tmp/seed
COPY --from=report /usr/local/static-report /tmp/analysis/report
# Login page assets are now embedded in the binary via go:embed

RUN echo "${VERSION}" > /etc/hub-build

# ENV DEVELOPMENT=1

RUN microdnf -y install \
  sqlite \
  openssh-clients \
  subversion \
  git \
  tar \
  vim \
 && microdnf -y clean all

RUN echo "hub:x:1001:0:hub:/:/sbin/nologin" >> /etc/passwd

ENTRYPOINT ["/usr/bin/tini", "--",  "/usr/local/bin/tackle-hub"]

LABEL name="konveyor/tackle2-hub" \
      description="Konveyor Tackle - Hub" \
      help="For more information visit https://konveyor.io" \
      license="Apache License 2.0" \
      maintainers="jortel@redhat.com,slucidi@redhat.com" \
      summary="Konveyor Tackle - Hub" \
      url="https://quay.io/repository/konveyor/tackle2-hub" \
      usage="podman run konveyor/tackle2-hub:latest" \
      com.redhat.component="konveyor-tackle-hub-container" \
      io.k8s.display-name="Tackle Hub" \
      io.k8s.description="Konveyor Tackle - Hub" \
      io.openshift.expose-services="" \
      io.openshift.tags="konveyor,tackle,hub" \
      io.openshift.min-cpu="100m" \
      io.openshift.min-memory="350Mi"
