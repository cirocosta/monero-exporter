ARG BUILDER_IMAGE=index.docker.io/library/ubuntu@sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3
ARG RUNTIME_IMAGE=gcr.io/distroless/base@sha256:97f9cbf81590c7dd878d7e10c20116fa7488c2fb91b3c839df1ebdcb61ab13ca


FROM $BUILDER_IMAGE AS builder

	ARG VERSION=8.0.6
	ARG SHA256=252a4d022f4dbd9a68d07cdd0e6762d62fa53462f294781cb53e8f3d61cd8649

	RUN set -ex && \
		apt update && \
		apt install -y curl gzip

	RUN set -ex && \
		curl -SOL https://dl.grafana.com/oss/release/grafana-${VERSION}.linux-amd64.tar.gz && \
		echo "${SHA256} grafana-${VERSION}.linux-amd64.tar.gz" | sha256sum -c && \
		tar xzf grafana-${VERSION}.linux-amd64.tar.gz --strip-components=1


FROM $RUNTIME_IMAGE

	USER nonroot:nonroot
	WORKDIR /grafana

	COPY --from=builder --chown=nonroot:nonroot /bin/grafana-server /usr/local/bin/grafana-server
	COPY --from=builder --chown=nonroot:nonroot /public             /grafana/public
	COPY --from=builder --chown=nonroot:nonroot /conf               /grafana/conf


	ENTRYPOINT [ "grafana-server" ]
