# Third-Party Licenses

This provider is built entirely on open-source dependencies. No dependency in the tree is
third-party proprietary code.

This list was generated from the actual compiled dependency graph (via
[`google/go-licenses`](https://github.com/google/go-licenses)), not just `go.mod` — it reflects
what actually ships in the release binary.

## Shipped in the release binary

| Module | License |
| --- | --- |
| `github.com/fatih/color` | MIT |
| `github.com/golang/protobuf/ptypes/empty` | BSD-3-Clause |
| `github.com/hashicorp/go-hclog` | MIT |
| `github.com/hashicorp/go-plugin` | MPL-2.0 |
| `github.com/hashicorp/go-uuid` | MPL-2.0 |
| `github.com/hashicorp/terraform-plugin-framework` | MPL-2.0 |
| `github.com/hashicorp/terraform-plugin-framework-validators` | MPL-2.0 |
| `github.com/hashicorp/terraform-plugin-go` | MPL-2.0 |
| `github.com/hashicorp/terraform-plugin-log` | MPL-2.0 |
| `github.com/hashicorp/terraform-registry-address` | MPL-2.0 |
| `github.com/hashicorp/terraform-svchost` | MPL-2.0 |
| `github.com/hashicorp/yamux` | MPL-2.0 |
| `github.com/mattn/go-colorable` | MIT |
| `github.com/mattn/go-isatty` | MIT |
| `github.com/mitchellh/go-testing-interface` | MIT |
| `github.com/oklog/run` | Apache-2.0 |
| `github.com/vmihailenco/msgpack/v5` | BSD-2-Clause |
| `github.com/vmihailenco/tagparser/v2` | BSD-2-Clause |
| `golang.org/x/net` | BSD-3-Clause |
| `golang.org/x/sys` | BSD-3-Clause |
| `golang.org/x/text` | BSD-3-Clause |
| `google.golang.org/genproto/googleapis/rpc` | Apache-2.0 |
| `google.golang.org/grpc` | Apache-2.0 |
| `google.golang.org/protobuf` | BSD-3-Clause |

## Build/test-time only (not shipped in the release binary)

| Module | License | Used for |
| --- | --- | --- |
| `github.com/hashicorp/terraform-plugin-testing` | MPL-2.0 | Acceptance test framework |
| `github.com/hashicorp/terraform-plugin-docs` | MPL-2.0 | Generating `docs/` via `tfplugindocs` |

## Summary

| License | Count |
| --- | --- |
| MPL-2.0 | 10 |
| MIT | 5 |
| BSD-3-Clause | 5 |
| Apache-2.0 | 3 |
| BSD-2-Clause | 2 |

All licenses are permissive open source. MPL-2.0 is weak-copyleft at the file level and does not
impose obligations on this project beyond redistributing MPL-licensed files under their original
license — it does not affect the licensing of this provider's own code (MPL-2.0, see [LICENSE](LICENSE)).

This provider itself is licensed under the [Mozilla Public License v2.0](LICENSE).

To regenerate this list, run `go-licenses csv .` (`go install github.com/google/go-licenses/v2@latest`).
