ARG BUILDER_IMAGE=index.docker.io/library/ubuntu@sha256:cf31af331f38d1d7158470e095b132acd126a7180a54f263d386da88eb681d93
ARG RUNTIME_IMAGE=gcr.io/distroless/base@sha256:bc84925113289d139a9ef2f309f0dd7ac46ea7b786f172ba9084ffdb4cbd9490


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
