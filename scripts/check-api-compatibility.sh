#!/usr/bin/env bash

set -euo pipefail

readonly module_path="github.com/uffejaeger/intercom-go"
readonly generated_package="${module_path}/internal/generated/intercom"
readonly baseline="${API_BASELINE:-v0.2.0}"
readonly apidiff_version="${APIDIFF_VERSION:-v0.0.0-20260727155853-b88d891fe743}"
readonly apidiff="golang.org/x/exp/cmd/apidiff@${apidiff_version}"
readonly repository_root="$(git rev-parse --show-toplevel)"
readonly work_dir="$(mktemp -d)"

cleanup() {
	rm -rf "${work_dir}"
}
trap cleanup EXIT

filter_reviewed_source_compatible_changes() {
	while IFS= read -r line; do
		case "${line}" in
			"- (*ContactsService).Create: changed from func(context.Context, ContactCreate) (*Contact, error) to func(context.Context, ContactCreate) (*Contact, error)" | \
			"- (*ContactsService).Update: changed from func(context.Context, string, ContactUpdate) (*Contact, error) to func(context.Context, string, ContactUpdate) (*Contact, error)" | \
			"- (*ContactIterator).Contact: changed from func() *Contact to func() *Contact" | \
			"- (*ContactsService).Get: changed from func(context.Context, string) (*Contact, error) to func(context.Context, string) (*Contact, error)" | \
			"- (*ContactsService).GetByExternalID: changed from func(context.Context, string) (*Contact, error) to func(context.Context, string) (*Contact, error)" | \
			"- (*ContactsService).List: changed from func(context.Context) (*ContactList, error) to func(context.Context) (*ContactList, error)" | \
			"- (*ContactsService).Merge: changed from func(context.Context, string, string) (*Contact, error) to func(context.Context, string, string) (*Contact, error)" | \
			"- (*ContactsService).Search: changed from func(context.Context, ContactSearch) (*ContactList, error) to func(context.Context, ContactSearch) (*ContactList, error)" | \
			"- (*CompaniesService).ListContacts: changed from func(context.Context, string) (*CompanyContacts, error) to func(context.Context, string) (*CompanyContacts, error)" | \
			"- (*ArticlesService).Create: changed from func(context.Context, ArticleCreate) (*Article, error) to func(context.Context, ArticleCreate) (*Article, error)" | \
			"- (*ArticlesService).List: changed from func(context.Context) (*ArticleList, error) to func(context.Context) (*ArticleList, error)" | \
			"- (*ArticlesService).Retrieve: changed from func(context.Context, string) (*Article, error) to func(context.Context, string) (*Article, error)" | \
			"- (*ArticlesService).Search: changed from func(context.Context, ArticleSearch) (*ArticleSearchResult, error) to func(context.Context, ArticleSearch) (*ArticleSearchResult, error)" | \
			"- (*ArticlesService).Update: changed from func(context.Context, string, ArticleUpdate) (*Article, error) to func(context.Context, string, ArticleUpdate) (*Article, error)" | \
			"- (*ConversationsService).ConvertToTicket: changed from func(context.Context, string, ConversationToTicket) (*Ticket, error) to func(context.Context, string, ConversationToTicket) (*Ticket, error)" | \
			"- (*TicketIterator).Ticket: changed from func() *Ticket to func() *Ticket" | \
			"- (*TicketsService).Create: changed from func(context.Context, TicketCreate) (*Ticket, error) to func(context.Context, TicketCreate) (*Ticket, error)" | \
			"- (*TicketsService).Get: changed from func(context.Context, string) (*Ticket, error) to func(context.Context, string) (*Ticket, error)" | \
			"- (*TicketsService).Search: changed from func(context.Context, TicketSearchQuery) (*TicketList, error) to func(context.Context, TicketSearchQuery) (*TicketList, error)" | \
			"- (*TicketsService).SearchWithOptions: changed from func(context.Context, TicketSearchQuery, CursorPageOptions) (*TicketList, error) to func(context.Context, TicketSearchQuery, CursorPageOptions) (*TicketList, error)" | \
			"- (*TicketsService).Update: changed from func(context.Context, string, TicketUpdate) (*Ticket, error) to func(context.Context, string, TicketUpdate) (*Ticket, error)" | \
			"- (*VisitorsService).Convert: changed from func(context.Context, VisitorConvert) (*VisitorConverted, error) to func(context.Context, VisitorConvert) (*VisitorConverted, error)" | \
			"- ContactCreate: changed from github.com/uffejaeger/intercom-go/internal/generated/intercom.CreateContactRequestSchema to ContactCreate" | \
			"- ContactUpdate: changed from github.com/uffejaeger/intercom-go/internal/generated/intercom.UpdateContactRequestSchema to ContactUpdate" | \
			"- Contact: changed from github.com/uffejaeger/intercom-go/internal/generated/intercom.ContactSchema to Contact" | \
			"- ContactList: changed from github.com/uffejaeger/intercom-go/internal/generated/intercom.ContactListSchema to ContactList" | \
			"- Article: changed from github.com/uffejaeger/intercom-go/internal/generated/intercom.ArticleListItemSchema to Article" | \
			"- ArticleList: changed from github.com/uffejaeger/intercom-go/internal/generated/intercom.ArticleListSchema to ArticleList" | \
			"- ArticleSearchResult: changed from github.com/uffejaeger/intercom-go/internal/generated/intercom.ArticleSearchResponseSchema to ArticleSearchResult" | \
			"- CompanyContacts: changed from github.com/uffejaeger/intercom-go/internal/generated/intercom.CompanyAttachedContactsSchema to CompanyContacts" | \
			"- Ticket: changed from github.com/uffejaeger/intercom-go/internal/generated/intercom.TicketSchema to Ticket" | \
			"- TicketList: changed from github.com/uffejaeger/intercom-go/internal/generated/intercom.TicketListSchema to TicketList" | \
			"- VisitorConverted: changed from github.com/uffejaeger/intercom-go/internal/generated/intercom.ContactSchema to Contact")
				;;
			*)
				printf '%s\n' "${line}"
				;;
		esac
	done
}

filter_reviewed_generated_alias_changes() {
	while IFS= read -r line; do
		case "${line}" in
			"- SearchRequestSchema: changed from SearchRequestSchema to SearchRequestSchema" | \
			"- SearchRequest_Query: changed from SearchRequest_Query to SearchRequest_Query")
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

	# API 2.16 changed the wire representation of Contact.OwnerId, Ticket
	# assignee IDs, and Article parent fields. The SDK uses hand-shaped boundary
	# models, including contact values nested under companies and visitor
	# conversion responses, to preserve the historical public field types while
	# converting the changed wire values.
	# apidiff reports each dependent method and iterator as a type-identity
	# change, so accept only these exact reviewed entries. External-consumer and
	# response-conversion regression tests verify the preserved source contract.
	if [[ "${label}" == "public" ]]; then
		report="$(printf '%s\n' "${report}" | filter_reviewed_source_compatible_changes)"
	elif [[ "${label}" == "generated-alias-model" ]]; then
		report="$(printf '%s\n' "${report}" | (
			cd "${repository_root}"
			go run ./internal/tools/filter-generated-api-diff "${repository_root}"
		) | filter_reviewed_generated_alias_changes)"
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
	go run "${apidiff}" -w "${work_dir}/baseline-generated.api" "${generated_package}"
)

echo "Exporting public API from the working tree..."
(
	cd "${repository_root}"
	go run "${apidiff}" -m -w "${work_dir}/current.api" "${module_path}"
	go run "${apidiff}" -w "${work_dir}/current-generated.api" "${generated_package}"
)

compare_api "public" "${work_dir}/baseline.api" "${work_dir}/current.api" -m
compare_api "generated-alias-model" \
	"${work_dir}/baseline-generated.api" "${work_dir}/current-generated.api" \
	-allow-internal

echo "Public API is backward compatible with ${baseline}."
