ARG SEED_ROOT=/opt/app-root/src/tackle2-seed
FROM registry.access.redhat.com/ubi9/go-toolset:latest as builder
ENV GOPATH=$APP_ROOT
COPY --chown=1001:0 . .
RUN make docker
ARG SEED_PROJECT=konveyor/tackle2-seed
ARG SEED_BRANCH=main
ARG SEED_ROOT
RUN if [ ! -d "${SEED_ROOT}" ]; then \
      git clone --branch ${SEED_BRANCH} https://github.com/${SEED_PROJECT} ${SEED_ROOT}; \
    fi

FROM quay.io/konveyor/static-report as report

FROM registry.access.redhat.com/ubi9/ubi-minimal
ARG SEED_ROOT
COPY --from=builder /opt/app-root/src/bin/hub /usr/local/bin/tackle-hub
COPY --from=builder /opt/app-root/src/auth/roles.yaml /tmp/roles.yaml
COPY --from=builder /opt/app-root/src/auth/users.yaml /tmp/users.yaml
COPY --from=builder ${SEED_ROOT}/resources/ /tmp/seed
COPY --from=report /usr/local/static-report /tmp/analysis/report

RUN microdnf -y install \
  sqlite \
 && microdnf -y clean all
ENTRYPOINT ["/usr/local/bin/tackle-hub"]

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
