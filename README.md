# monero-exporter

[Prometheus](https://prometheus.io) exporter for [monero](https://getmonero.org).


## Installation

Using Go:

```
go get github.com/cirocosta/monero-exporter/cmd/monero-exporter
```

From releases:

**TODO**


## Usage

```console
$ monero-exporter --help

Prometheus exporter for monero metrics

Usage:
  monero-exporter [flags]
  monero-exporter [command]

Available Commands:
  completion  generate the autocompletion script for the specified shell
  help        Help about any command
  version     print the version of this CLI

Flags:
      --address string          address of the monero node to collect metrics from
      --geoip-filepath string   filepath of a geoip database file for ip to country resolution
  -h, --help                    help for monero-exporter

Use "monero-exporter [command] --help" for more information about a command.
```

## Metrics


| name | description |
| ---- | ----------- |
| monero_bans | number of nodes banned |
| monero_connections | connections info |
| monero_connections_livetime | peers livetime distribution |
| monero_fee_estimate | fee estimate for 1 grace block |
| monero_height_divergence | how much our peers diverge from us in block height |
| monero_info_alt_blocks_count | info for alt_blocks_count |
| monero_info_block_size_limit | info for block_size_limit |
| monero_info_block_size_median | info for block_size_median |
| monero_info_busy_syncing | info for busy_syncing |
| monero_info_cumulative_difficulty | info for cumulative_difficulty |
| monero_info_difficulty | info for difficulty |
| monero_info_free_space | info for free_space |
| monero_info_grey_peerlist_size | info for grey_peerlist_size |
| monero_info_height | info for height |
| monero_info_height_without_bootstrap | info for height_without_bootstrap |
| monero_info_incoming_connections_count | info for incoming_connections_count |
| monero_info_mainnet | info for mainnet |
| monero_info_offline | info for offline |
| monero_info_outgoing_connections_count | info for outgoing_connections_count |
| monero_info_rpc_connections_count | info for rpc_connections_count |
| monero_info_stagenet | info for stagenet |
| monero_info_start_time | info for start_time |
| monero_info_synchronized | info for synchronized |
| monero_info_target | info for target |
| monero_info_target_height | info for target_height |
| monero_info_testnet | info for testnet |
| monero_info_tx_count | info for tx_count |
| monero_info_tx_pool_size | info for tx_pool_size |
| monero_info_untrusted | info for untrusted |
| monero_info_was_bootstrap_ever_used | info for was_bootstrap_ever_used |
| monero_info_white_peerlist_size | info for white_peerlist_size |
| monero_last_block_header_block_size | info for block_size |
| monero_last_block_header_block_weight | info for block_weight |
| monero_last_block_header_cumulative_difficulty | info for cumulative_difficulty |
| monero_last_block_header_cumulative_difficulty_top64 | info for cumulative_difficulty_top64 |
| monero_last_block_header_depth | info for depth |
| monero_last_block_header_difficulty | info for difficulty |
| monero_last_block_header_difficulty_top64 | info for difficulty_top64 |
| monero_last_block_header_height | info for height |
| monero_last_block_header_long_term_weight | info for long_term_weight |
| monero_last_block_header_major_version | info for major_version |
| monero_last_block_header_minor_version | info for minor_version |
| monero_last_block_header_nonce | info for nonce |
| monero_last_block_header_num_txes | info for num_txes |
| monero_last_block_header_orphan_status | info for orphan_status |
| monero_last_block_header_reward | info for reward |
| monero_last_block_header_timestamp | info for timestamp |
| monero_last_block_txn_fee | distribution of outputs in last block |
| monero_last_block_txn_size | distribution of tx sizes |
| monero_last_block_vin | distribution of inputs in last block |
| monero_last_block_vout | distribution of outputs in last block |
| monero_mempool_bytes_max | info for bytes_max |
| monero_mempool_bytes_med | info for bytes_med |
| monero_mempool_bytes_min | info for bytes_min |
| monero_mempool_bytes_total | info for bytes_total |
| monero_mempool_fee_total | info for fee_total |
| monero_mempool_histo_98pc | info for histo_98pc |
| monero_mempool_num_10m | info for num_10m |
| monero_mempool_num_double_spends | info for num_double_spends |
| monero_mempool_num_failing | info for num_failing |
| monero_mempool_num_not_relayed | info for num_not_relayed |
| monero_mempool_oldest | info for oldest |
| monero_mempool_txs_total | info for txs_total |
| monero_net_total_in_bytes | network statistics |
| monero_net_total_out_bytes | network statistics |
| monero_peers_new | peers info |
| monero_rpc_count | todo |
| monero_rpc_time | todo |


## Donate

![xmr address](./assets/donate.png)

891B5keCnwXN14hA9FoAzGFtaWmcuLjTDT5aRTp65juBLkbNpEhLNfgcBn6aWdGuBqBnSThqMPsGRjWVQadCrhoAT6CnSL3
