#!/usr/bin/env bash

set -euo pipefail

readonly module_path="github.com/uffejaeger/intercom-go"
readonly baseline="${API_BASELINE:-v0.2.0}"
readonly apidiff_version="${APIDIFF_VERSION:-v0.0.0-20260727155853-b88d891fe743}"
readonly apidiff="golang.org/x/exp/cmd/apidiff@${apidiff_version}"
readonly repository_root="$(git rev-parse --show-toplevel)"
readonly work_dir="$(mktemp -d)"

cleanup() {
	rm -rf "${work_dir}"
}
trap cleanup EXIT

filter_source_compatible_contact_request_changes() {
	while IFS= read -r line; do
		case "${line}" in
			"- (*ContactsService).Create: changed from func(context.Context, ContactCreate) (*Contact, error) to func(context.Context, ContactCreate) (*Contact, error)" | \
			"- (*ContactsService).Update: changed from func(context.Context, string, ContactUpdate) (*Contact, error) to func(context.Context, string, ContactUpdate) (*Contact, error)" | \
			"- ContactCreate: changed from github.com/uffejaeger/intercom-go/internal/generated/intercom.CreateContactRequestSchema to ContactCreate" | \
			"- ContactUpdate: changed from github.com/uffejaeger/intercom-go/internal/generated/intercom.UpdateContactRequestSchema to ContactUpdate")
				;;
			*)
				printf '%s\n' "${line}"
				;;
		esac
	done
}

compare_api() {
	local label="$1"
	local old_export="$2"
	local new_export="$3"
	local comparison_errors="${work_dir}/${label}.stderr"
	local report
	shift 3

	if ! report="$(go run "${apidiff}" "$@" -incompatible \
		"${old_export}" "${new_export}" 2>"${comparison_errors}")"; then
		cat "${comparison_errors}" >&2
		echo "apidiff could not compare the ${label} APIs." >&2
		exit 1
	fi

	# ContactCreate and ContactUpdate changed from aliases of generated models to
	# hand-shaped SDK requests in API 2.16. This intentionally preserves the
	# importable, historical OwnerId *int source contract while converting to the
	# upstream string wire format. apidiff reports the required type-identity
	# change as incompatible, so accept only these exact entries; the external
	# consumer regression test verifies the preserved source contract.
	if [[ "${label}" == "public" ]]; then
		report="$(printf '%s\n' "${report}" | filter_source_compatible_contact_request_changes)"
	fi

	sed '/^Ignoring internal package /d' "${comparison_errors}" >&2

	if [[ -n "${report}" ]]; then
		echo "Backward-incompatible ${label} API changes detected against ${baseline}:" >&2
		echo "${report}" >&2
		exit 1
	fi
}

if ! git -C "${repository_root}" cat-file -e "${baseline}^{commit}" 2>/dev/null; then
	echo "API baseline ${baseline} is unavailable." >&2
	echo "Fetch release tags with: git fetch --tags origin" >&2
	exit 1
fi

mkdir -p "${work_dir}/baseline"
git -C "${repository_root}" archive "${baseline}" | tar -x -C "${work_dir}/baseline"

echo "Exporting public API from ${baseline}..."
(
	cd "${work_dir}/baseline"
	go run "${apidiff}" -m -w "${work_dir}/baseline.api" "${module_path}"
)

echo "Exporting public API from the working tree..."
(
	cd "${repository_root}"
	go run "${apidiff}" -m -w "${work_dir}/current.api" "${module_path}"
)

compare_api "public" "${work_dir}/baseline.api" "${work_dir}/current.api" -m

echo "Public API is backward compatible with ${baseline}."
