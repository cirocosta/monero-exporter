ARG BUILDER_IMAGE=index.docker.io/library/ubuntu@sha256:cf31af331f38d1d7158470e095b132acd126a7180a54f263d386da88eb681d93
ARG RUNTIME_IMAGE=gcr.io/distroless/base@sha256:bc84925113289d139a9ef2f309f0dd7ac46ea7b786f172ba9084ffdb4cbd9490


FROM $BUILDER_IMAGE AS builder

	ARG VERSION=2.28.1
	ARG SHA256=91dd91e13f30fe520e01175ca1027dd09a458d4421a584ba557ba88b38803f27

	RUN set -ex && \
		apt update && \
		apt install -y curl gzip

	RUN set -ex && \
		curl -SOL https://github.com/prometheus/prometheus/releases/download/v${VERSION}/prometheus-${VERSION}.linux-amd64.tar.gz && \
		echo "${SHA256} prometheus-${VERSION}.linux-amd64.tar.gz" | sha256sum -c && \
		tar xvzf prometheus-${VERSION}.linux-amd64.tar.gz --strip-components=1 && \
		mv ./prometheus /usr/local/bin/prometheus

	RUN mkdir /data


FROM $RUNTIME_IMAGE

	COPY --from=builder --chown=nonroot:nonroot /data 		      /data
	COPY --from=builder --chown=nonroot:nonroot /usr/local/bin/prometheus /usr/local/bin/prometheus
	USER nonroot:nonroot

	ENTRYPOINT [ "prometheus" ]
