#!/bin/sh
#
# Environment variables:
#
#   DNSIMPLE_INSTALL   - installation directory (default: $HOME/.dnsimple)
#   DNSIMPLE_BASE_URL  - base URL for downloading releases
#
# Examples:
#
#   curl -fsSL https://raw.githubusercontent.com/dnsimple/dnsimple-cli/main/install.sh | sh
#   curl -fsSL .../install.sh | sh -s -- --base-url https://github.com/dnsimple/cli/releases

set -e

BOLD="$(tput bold 2>/dev/null || printf '')"
RED="$(tput setaf 1 2>/dev/null || printf '')"
GREEN="$(tput setaf 2 2>/dev/null || printf '')"
YELLOW="$(tput setaf 3 2>/dev/null || printf '')"
BLUE="$(tput setaf 4 2>/dev/null || printf '')"
NO_COLOR="$(tput sgr0 2>/dev/null || printf '')"

DEFAULT_BASE_URL="https://github.com/dnsimple/dnsimple-cli-public/releases"
BINARY_NAME="dnsimple"

info() {
	printf '%s\n' "${BOLD}>${NO_COLOR} $*"
}

warn() {
	printf '%s\n' "${YELLOW}! $*${NO_COLOR}"
}

error() {
	printf '%s\n' "${RED}x $*${NO_COLOR}" >&2
}

has() {
	command -v "$1" 1>/dev/null 2>&1
}

download() {
	file="$1"
	url="$2"

	if has curl; then
		cmd="curl --fail --silent --location --output $file $url"
	elif has wget; then
		cmd="wget --quiet --output-document=$file $url"
	else
		error "No HTTP download program (curl, wget) found."
		return 1
	fi

	$cmd && return 0 || rc=$?

	error "Download failed (exit code $rc): ${BLUE}${url}${NO_COLOR}"
	return $rc
}

detect_os() {
	os="$(uname -s | tr '[:upper:]' '[:lower:]')"

	case "${os}" in
	linux) os="linux" ;;
	darwin) os="darwin" ;;
	mingw* | msys* | cygwin*) os="windows" ;;
	*)
		error "Unsupported operating system: ${os}"
		exit 1
		;;
	esac

	printf '%s' "${os}"
}

detect_arch() {
	arch="$(uname -m | tr '[:upper:]' '[:lower:]')"

	case "${arch}" in
	x86_64 | amd64) arch="amd64" ;;
	aarch64 | arm64) arch="arm64" ;;
	*)
		error "Unsupported architecture: ${arch}"
		exit 1
		;;
	esac

	printf '%s' "${arch}"
}

detect_version() {
	version="$1"
	base_url="$2"

	if [ -n "${version}" ]; then
		# Strip leading 'v' if present, we add it back as needed
		version="${version#v}"
		printf '%s' "${version}"
		return
	fi

	info "Fetching latest version..."

	# Follow the /latest redirect and extract the version from the final URL
	if has curl; then
		redirect_url="$(curl -sI -o /dev/null -w '%{url_effective}' -L "${base_url}/latest")"
	elif has wget; then
		redirect_url="$(wget --spider -S -O /dev/null "${base_url}/latest" 2>&1 | grep -i 'Location:' | tail -1 | awk '{print $2}' | tr -d '\r')"
	fi

	version="$(printf '%s' "${redirect_url:-}" | sed -E 's|.*/v?([0-9][^/]*)$|\1|')"

	if [ -z "${version}" ]; then
		error "Failed to determine latest version."
		error "Please specify a version: install.sh --version 0.1.0"
		exit 1
	fi

	printf '%s' "${version}"
}

verify_checksum() {
	file="$1"
	expected_file="$2"
	archive_name="$3"

	if has sha256sum; then
		actual="$(sha256sum "${file}" | awk '{print $1}')"
	elif has shasum; then
		actual="$(shasum -a 256 "${file}" | awk '{print $1}')"
	else
		warn "No checksum program found (sha256sum, shasum). Skipping verification."
		return 0
	fi

	expected="$(grep "${archive_name}" "${expected_file}" | awk '{print $1}')"

	if [ -z "${expected}" ]; then
		warn "No checksum found for ${archive_name}. Skipping verification."
		return 0
	fi

	if [ "${actual}" != "${expected}" ]; then
		error "Checksum verification failed!"
		error "  Expected: ${expected}"
		error "  Actual:   ${actual}"
		return 1
	fi
}

trusty_says() {
	case $(awk 'BEGIN{srand(); printf "%d", rand()*6}') in
	0) printf '%s' "Your CLI has resolved successfully!" ;;
	1) printf '%s' "Ready to serve your domains!" ;;
	2) printf '%s' "Hi, I'm Trusty! Let's manage some domains." ;;
	3) printf '%s' "All records point to success!" ;;
	4) printf '%s' "You're all set! Let's make DNS simple." ;;
	5) printf '%s' "DNS made simple, right from your terminal!" ;;
	esac
}

print_banner() {
	_version="$1"
	_message="$(trusty_says)"
	printf '\n'
	printf '%s             .----.%s\n' "${YELLOW}" "${NO_COLOR}"
	printf '%s        .----|    |----.%s\n' "${YELLOW}" "${NO_COLOR}"
	printf '%s        |              |%s\n' "${YELLOW}" "${NO_COLOR}"
	printf '%s        |  (O)    (O)  |%s   %s%sDNSimple CLI v%s installed%s\n' "${YELLOW}" "${NO_COLOR}" "${GREEN}" "${BOLD}" "${_version}" "${NO_COLOR}"
	printf '%s        |              |%s\n' "${YELLOW}" "${NO_COLOR}"
	printf '%s        |   \\______/   |%s   "%s"\n' "${YELLOW}" "${NO_COLOR}" "${_message}"
	printf '%s        |              |%s\n' "${YELLOW}" "${NO_COLOR}"
	printf '%s        |     [*]      |%s   Run %sdnsimple --help%s to get started.\n' "${YELLOW}" "${NO_COLOR}" "${BOLD}" "${NO_COLOR}"
	printf '%s%s%s\n' "${YELLOW}" "        '--------------'" "${NO_COLOR}"
	printf '\n'
}

print_help() {
	printf '%s\n' \
		"Install script for the DNSimple CLI" \
		"" \
		"Usage:" \
		"  install.sh [options]" \
		"" \
		"Options:" \
		"  -v, --version <version>   Install a specific version (default: latest)" \
		"  -d, --dir <path>          Installation directory (default: \$HOME/.dnsimple)" \
		"  -b, --base-url <url>      Base URL for downloading releases" \
		"  -y, --yes                 Skip confirmation prompt" \
		"  -h, --help                Print this help message"
}

main() {
	version=""
	install_dir=""
	base_url=""
	force=false

	while [ "$#" -gt 0 ]; do
		case "$1" in
		-v | --version)
			version="$2"
			shift 2
			;;
		-v=* | --version=*)
			version="${1#*=}"
			shift 1
			;;
		-d | --dir)
			install_dir="$2"
			shift 2
			;;
		-d=* | --dir=*)
			install_dir="${1#*=}"
			shift 1
			;;
		-b | --base-url)
			base_url="$2"
			shift 2
			;;
		-b=* | --base-url=*)
			base_url="${1#*=}"
			shift 1
			;;
		-y | --yes)
			force=true
			shift 1
			;;
		-h | --help)
			print_help
			exit 0
			;;
		*)
			error "Unknown option: $1"
			print_help
			exit 1
			;;
		esac
	done

	os="$(detect_os)"
	arch="$(detect_arch)"
	base_url="${base_url:-${DNSIMPLE_BASE_URL:-${DEFAULT_BASE_URL}}}"
	# Strip trailing slash
	base_url="${base_url%/}"
	version="$(detect_version "${version}" "${base_url}")"

	install_dir="${install_dir:-${DNSIMPLE_INSTALL:-$HOME/.dnsimple}}"
	bin_dir="${install_dir}/bin"

	ext="tar.gz"
	if [ "${os}" = "windows" ]; then
		ext="zip"
	fi

	archive_name="${BINARY_NAME}_${version}_${os}_${arch}.${ext}"
	download_url="${base_url}/download/v${version}/${archive_name}"
	checksums_url="${base_url}/download/v${version}/checksums.txt"

	printf '\n'
	info "${BOLD}Version${NO_COLOR}:   ${GREEN}${version}${NO_COLOR}"
	info "${BOLD}Platform${NO_COLOR}:  ${GREEN}${os}/${arch}${NO_COLOR}"
	info "${BOLD}Location${NO_COLOR}:  ${GREEN}${bin_dir}${NO_COLOR}"
	printf '\n'

	if [ "${force}" = false ] && [ -t 0 ]; then
		printf '%s ' "Install DNSimple CLI ${GREEN}v${version}${NO_COLOR}? ${BOLD}[y/N]${NO_COLOR}"
		read -r yn </dev/tty
		case "${yn}" in
		[yY] | [yY][eE][sS]) ;;
		*)
			error "Installation cancelled."
			exit 1
			;;
		esac
	fi

	tmp_dir="$(mktemp -d)"
	trap 'rm -rf "${tmp_dir}"' EXIT

	info "Downloading ${BLUE}${archive_name}${NO_COLOR}..."
	download "${tmp_dir}/${archive_name}" "${download_url}"

	info "Verifying checksum..."
	download "${tmp_dir}/checksums.txt" "${checksums_url}"
	verify_checksum "${tmp_dir}/${archive_name}" "${tmp_dir}/checksums.txt" "${archive_name}"

	info "Extracting..."
	mkdir -p "${bin_dir}"
	if [ "${ext}" = "tar.gz" ]; then
		tar -xzf "${tmp_dir}/${archive_name}" -C "${tmp_dir}"
	else
		if has unzip; then
			unzip -o "${tmp_dir}/${archive_name}" -d "${tmp_dir}" >/dev/null
		elif has 7z; then
			7z x -o"${tmp_dir}" -y "${tmp_dir}/${archive_name}" >/dev/null
		else
			error "No unzip program found (unzip, 7z)."
			exit 1
		fi
	fi

	# Move binary into place
	mv "${tmp_dir}/${BINARY_NAME}" "${bin_dir}/${BINARY_NAME}"
	chmod +x "${bin_dir}/${BINARY_NAME}"

	print_banner "${version}"

	if command -v "${BINARY_NAME}" >/dev/null 2>&1; then
		printf '\n'
		info "Run '${BOLD}dnsimple --help${NO_COLOR}' to get started."
		return
	fi

	printf '\n'

	# Determine the appropriate shell config file
	case "${SHELL:-}" in
	*/zsh)
		shell_config="$HOME/.zshrc"
		shell_name="~/.zshrc"
		;;
	*/bash)
		if [ -f "$HOME/.bashrc" ]; then
			shell_config="$HOME/.bashrc"
			shell_name="~/.bashrc"
		else
			shell_config="$HOME/.bash_profile"
			shell_name="~/.bash_profile"
		fi
		;;
	*/fish)
		shell_config="$HOME/.config/fish/config.fish"
		shell_name="~/.config/fish/config.fish"
		;;
	*)
		shell_config=""
		shell_name=""
		;;
	esac

	export_line="export PATH=\"${bin_dir}:\$PATH\""
	if [ "${SHELL:-}" = "*/fish" ]; then
		export_line="set -gx PATH \"${bin_dir}\" \$PATH"
	fi

	# Check if already in a config file
	if [ -n "${shell_config}" ] && [ -f "${shell_config}" ] && grep -q "${bin_dir}" "${shell_config}" 2>/dev/null; then
		info "PATH is already configured in ${BOLD}${shell_name}${NO_COLOR}"
		info "Restart your shell or run: ${BOLD}source ${shell_name}${NO_COLOR}"
		return
	fi

	if [ -n "${shell_config}" ] && { [ "${force}" = true ] || [ ! -t 0 ]; }; then
		# Non-interactive or --yes: don't modify PATH automatically
		warn "Add the following to ${BOLD}${shell_name}${NO_COLOR}:"
		printf '\n  %s\n\n' "${export_line}"
		return
	fi

	if [ -n "${shell_config}" ] && [ -t 0 ]; then
		printf '%s ' "Add to ${BOLD}${shell_name}${NO_COLOR}? ${BOLD}[y/N]${NO_COLOR}"
		read -r yn </dev/tty
		case "${yn}" in
		[yY] | [yY][eE][sS])
			printf '\n# DNSimple CLI\n%s\n' "${export_line}" >>"${shell_config}"
			info "PATH configured in ${BOLD}${shell_name}${NO_COLOR}"
			info "Restart your shell or run: ${BOLD}source ${shell_name}${NO_COLOR}"
			return
			;;
		esac
	fi

	warn "Manually add to your shell profile:"
	printf '\n  %s\n\n' "${export_line}"
}

main "$@"
