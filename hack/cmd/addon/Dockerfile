FROM registry.access.redhat.com/ubi8/go-toolset:1.16.7 as builder
ENV GOPATH=$APP_ROOT
COPY . .
RUN make addon

FROM registry.access.redhat.com/ubi8/ubi-minimal:8.4
RUN microdnf -y install findutils && microdnf clean all
COPY --from=builder /opt/app-root/src/bin/addon /usr/local/bin/addon
ENTRYPOINT ["/usr/local/bin/addon"]

LABEL name="konveyor/addon" \
      description="Konveyor Tackle - Test Addon" \
      help="For more information visit https://konveyor.io" \
      license="Apache License 2.0" \
      maintainers="jortel@redhat.com" \
      summary="Konveyor Tackle - Test Addon" \
      url="https://quay.io/repository/konveyor/addon" \
      usage="podman run konveyor/addon:latest" \
      com.redhat.component="konveyor-tackle-addon" \
      io.k8s.display-name="Tackle Addon" \
      io.k8s.description="Konveyor Tackle - Addon" \
      io.openshift.expose-services="" \
      io.openshift.tags="konveyor,tackle,addon" \
      io.openshift.min-cpu="100m" \
      io.openshift.min-memory="350Mi"
