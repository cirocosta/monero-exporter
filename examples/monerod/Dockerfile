ARG BUILDER_IMAGE=index.docker.io/library/ubuntu@sha256:82becede498899ec668628e7cb0ad87b6e1c371cb8a1e597d83a47fac21d6af3
ARG RUNTIME_IMAGE=gcr.io/distroless/base@sha256:97f9cbf81590c7dd878d7e10c20116fa7488c2fb91b3c839df1ebdcb61ab13ca


FROM $BUILDER_IMAGE AS builder

	ARG MONERO_VERSION=0.17.2.0
	ARG MONERO_SHA256=59e16c53b2aff8d9ab7a8ba3279ee826ac1f2480fbb98e79a149e6be23dd9086

	RUN set -ex && \
		apt update && \
		apt install -y curl bzip2

	RUN set -ex && \
		curl -SOL https://downloads.getmonero.org/cli/monero-linux-x64-v${MONERO_VERSION}.tar.bz2 && \
		echo "${MONERO_SHA256} monero-linux-x64-v${MONERO_VERSION}.tar.bz2" | sha256sum -c && \
		tar xf monero-linux-x64-v${MONERO_VERSION}.tar.bz2 --strip-components=1 && \
		mv ./monerod ./monero-wallet-rpc /usr/local/bin

	RUN mkdir /data


FROM $RUNTIME_IMAGE

	COPY --from=builder --chown=nonroot:nonroot /data 		      		/data
	COPY --from=builder --chown=nonroot:nonroot /usr/local/bin/monerod 		/usr/local/bin/monerod
	COPY --from=builder --chown=nonroot:nonroot /usr/local/bin/monero-wallet-rpc 	/usr/local/bin/monero-wallet-rpc
	USER nonroot:nonroot

	ENTRYPOINT [ "monerod", "--non-interactive" ]
