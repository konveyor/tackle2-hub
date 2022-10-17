FROM registry.access.redhat.com/ubi9/go-toolset:latest as builder
ENV GOPATH=$APP_ROOT
COPY --chown=1001:0 . .
RUN make docker

FROM registry.access.redhat.com/ubi9/ubi-minimal
COPY --from=builder /opt/app-root/src/bin/hub /usr/local/bin/tackle-hub
COPY --from=builder /opt/app-root/src/auth/roles.yaml /tmp/roles.yaml
COPY --from=builder /opt/app-root/src/auth/users.yaml /tmp/users.yaml
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
