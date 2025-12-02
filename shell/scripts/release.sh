#!/usr/bin/env bash
set -euo pipefail



########################################
# Config
########################################

# Parsed CLI options
TYPE=""       # major|minor|patch
FORCE=0       # 1 = skip confirmation
MESSAGE=""    # currently unused, kept for parity

########################################
# Helpers
########################################

usage() {
  cat <<EOF
Usage: $(basename "$0") [major|minor|patch] [--force] [--message "text"]

Arguments:
  major|minor|patch   Optional release type. If omitted, it will be detected
                      from git diff (like your Laravel command).
Options:
  --force             Don't ask confirmation before creating the tag.
EOF
}

# Simple info/warn helpers
info()  { printf '==> %s\n' "$*"; }
warn()  { printf '!!  %s\n' "$*" >&2; }
replace_last_line() {
  # Move cursor 1 line up and clear the whole line
  printf '\033[1A\033[2K'
  printf '%s\n' "==> $*"
}

########################################
# Parse arguments
########################################

while [[ $# -gt 0 ]]; do
  case "$1" in
    major|minor|patch)
      TYPE="$1"
      shift
      ;;
    --force)
      FORCE=1
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      warn "Unknown argument: $1"
      usage
      exit 1
      ;;
  esac
done

########################################
# Move to script directory, load .env
########################################

# Load .env if present and GITHUB_TOKEN not already set
if [[ -f .env ]]; then
  set -a
  # shellcheck disable=SC1091
  source .env
  set +a
else
  warn "No .env file is present"
fi

# Directory containing the git repo (equivalent to "cd src" in your PHP code)
SRC_DIR="${SRC_DIR:-src}"
GIT_DIR="${GIT_DIR:-$SRC_DIR}"

# Validate GIT_DIR
if ! git -C "$GIT_DIR" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  warn "GIT_DIR '$GIT_DIR' is not a git working tree"
  exit 1
fi

if [[ -z "${GITHUB_TOKEN:-}" ]]; then
  warn "GITHUB_TOKEN environment variable (or .env entry) is required"
  exit 1
fi

# Check dependencies
for cmd in git curl jq; do
  if ! command -v "$cmd" >/dev/null 2>&1; then
    warn "Required command '$cmd' not found in PATH"
    exit 1
  fi
done

########################################
# Globals populated by functions
########################################

GITHUB_REPOSITORY=""   # "owner/repo"
OLD_TAG=""             # e.g. "v1.2.3"
OLD_VERSION=""         # e.g. "1.2.3"
NEW_VERSION=""         # e.g. "1.3.0"
NEW_TAG=""             # e.g. "v1.3.0"
CHANGES=""             # release body
RELEASE_URL=""         # html_url from GitHub

########################################
# Functions (port of your PHP logic)
########################################

check_uncommitted_changes() {
  local label="Checking for uncommitted changes";
  info "$label...";
  local status
  if ! status=$(cd "$GIT_DIR" && git status --porcelain); then
    replace_last_line "$label ✖"
    warn "Failed to check for uncommitted changes"
    return 1
  fi

  if [[ -n "$status" ]]; then
    replace_last_line "$label ⚠"
    warn "⚠️  There are uncommitted changes in your working directory:"
    printf '%s\n' "$status"
    return 1
  fi
  
  replace_last_line "$label ✔"
}

get_repository() {
  local label="Detecting repository"
  info "$label..."
  local url repo
  if ! url=$(cd "$GIT_DIR" && git config --get remote.origin.url); then
    replace_last_line "$label ⚠"
    warn "Failed to detect repository (git config remote.origin.url)"
    return 1
  fi
  url=$(printf '%s' "$url" | tr -d '\n' | tr -d '\r')

  if [[ "$url" == *"github.com:"* ]]; then
    # git@github.com:owner/repo.git
    repo="${url##*:}"
  elif [[ "$url" == *"github.com/"* ]]; then
    # https://github.com/owner/repo.git
    repo="${url#*github.com/}"
  else
    replace_last_line "$label ⚠"
    warn "Remote origin does not look like a GitHub URL: $url"
    return 1
  fi

  repo="${repo%.git}"
  GITHUB_REPOSITORY="$repo"
  replace_last_line "$label: $GITHUB_REPOSITORY ✔"
}

get_current_version() {
  local label="Fetching latest GitHub release"
  info "$label..."
  local resp tag
  if ! resp=$(
    curl -sS \
      -H "Authorization: Bearer $GITHUB_TOKEN" \
      -H "Accept: application/vnd.github+json" \
      "https://api.github.com/repos/$GITHUB_REPOSITORY/releases/latest"
  ); then
    replace_last_line "$label ⚠"
    warn "Failed to call GitHub API for latest release"
    return 1
  fi

  tag=$(jq -r '.tag_name // empty' <<<"$resp" || true)
  if [[ -z "$tag" || "$tag" == "null" ]]; then
    replace_last_line "$label ⚠"
    warn "Failed to fetch latest tag from GitHub"
    printf '%s\n' "$resp"
    return 1
  fi

  OLD_TAG="$tag"
  OLD_VERSION="${tag#v}"

  replace_last_line "$label: $OLD_TAG ✔"
}

get_php_version_from_composer() {
  # Non-fatal; returns nothing on failure.
  local path="$SRC_DIR/composer.json"
  if [[ ! -f "$path" ]]; then
    warn "composer.json not found at $path"
    return 0
  fi

  local constraint
  constraint=$(jq -r '.require.php // empty' "$path" 2>/dev/null || true)

  if [[ -z "$constraint" || "$constraint" == "null" ]]; then
    warn "No PHP version specified in composer.json"
    return 0
  fi

  if [[ "$constraint" =~ ([0-9]+\.[0-9]+) ]]; then
    printf '%s\n' "${BASH_REMATCH[1]}"
  fi
}

detect_release_type() {
  # If user provided the type, respect it.
  if [[ -n "${TYPE:-}" ]]; then
    info "Using provided release type: $TYPE"
    return 0
  fi

  info "Detecting release type from git diff since $OLD_TAG..."

  local changed
  git -C "$GIT_DIR" fetch --tags
  if ! changed=$(cd "$GIT_DIR" && git diff --name-only "$OLD_TAG..HEAD"); then
    warn "Failed to run git diff for changed files"
    return 1
  fi

  if [[ -z "$changed" ]]; then
    info "No code changes detected → patch"
    TYPE="patch"
    return 0
  fi

  local phpFiles=() migrations=() tests=() docs=() configs=() views=() composerFiles=()

  while IFS= read -r file; do
    [[ -z "$file" ]] && continue
    if [[ "$file" == *.php ]]; then phpFiles+=("$file"); fi
    if [[ "$file" == database/migrations/* ]]; then migrations+=("$file"); fi
    if [[ "$file" == tests/* ]]; then tests+=("$file"); fi
    if [[ "$file" == docs/* || "$file" =~ \.(md|rst|txt)$ ]]; then docs+=("$file"); fi
    if [[ "$file" == config/* ]]; then configs+=("$file"); fi
    if [[ "$file" == resources/views* ]]; then views+=("$file"); fi
    if [[ "$file" == "composer.json" || "$file" == "composer.lock" ]]; then composerFiles+=("$file"); fi
  done <<<"$changed"

  local phpVersion
  phpVersion=$(get_php_version_from_composer || true)
  phpVersion="${phpVersion:-8.2}"
  info "🔍 Using PHP $phpVersion for parsing changes from $OLD_TAG"

  local major=0 minor=0

  # Analyse PHP diffs (port of your heuristics)
  for file in "${phpFiles[@]}"; do
    local diff
    diff=$(cd "$GIT_DIR" && git diff "$OLD_TAG..HEAD" -- "$file") || diff=""

    local removed=() added=()
    while IFS= read -r line; do
      case "$line" in
        -*)
          removed+=("${line#-}")
          ;;
        +*)
          added+=("${line#+}")
          ;;
      esac
    done <<<"$diff"

    # Added lines
    for line in "${added[@]}"; do
      if grep -Eq '\b(class|interface|trait|enum)[[:space:]]+[A-Za-z0-9_]+' <<<"$line"; then
        info "- added class/trait/interface/enum → Minor [$line]"
        minor=1
      fi
      if grep -Eq 'public[[:space:]]+function[[:space:]]+[A-Za-z0-9_]+[[:space:]]*\(' <<<"$line"; then
        info "- added public method → Minor [$line]"
        minor=1
      fi
      if grep -Eq 'public[[:space:]]+\$[A-Za-z0-9_]+' <<<"$line"; then
        info "- added public property → Minor [$line]"
        minor=1
      fi
      if grep -Eq 'public[[:space:]]+const[[:space:]]+[A-Za-z0-9_]+' <<<"$line"; then
        info "- added public constant → Minor [$line]"
        minor=1
      fi
    done

    # Removed / changed lines
    for line in "${removed[@]}"; do
      if grep -Eq '\b(class|interface|trait|enum)[[:space:]]+[A-Za-z0-9_]+' <<<"$line"; then
        info "- removed class/trait/interface/enum → MAJOR [$line]"
        major=1
      fi
      if grep -Eq 'public[[:space:]]+function[[:space:]]+[A-Za-z0-9_]+[[:space:]]*\(' <<<"$line"; then
        info "- removed public method → MAJOR [$line]"
        major=1
      fi

      # Changed visibility
      if [[ "$line" =~ (public|protected|private)[[:space:]]+function[[:space:]]+([A-Za-z0-9_]+) ]]; then
        local vis1="${BASH_REMATCH[1]}"
        local fname="${BASH_REMATCH[2]}"
        for a in "${added[@]}"; do
          if [[ "$a" =~ (public|protected|private)[[:space:]]+function[[:space:]]+$fname ]]; then
            local vis2="${BASH_REMATCH[1]}"
            if [[ "$vis1" != "$vis2" ]]; then
              info "- visibility changed for $fname → MAJOR [$line]"
              major=1
            fi
          fi
        done
      fi

      # Changed return type
      local re_params_fn_name='public[[:space:]]+function[[:space:]]+([A-Za-z0-9_]+)[[:space:]]*\\((.*?)\\)'
      local re_params_same_name_prefix="public[[:space:]]+function[[:space:]]+"

      if [[ "$line" =~ $re_params_fn_name ]]; then
        local fname_r="${BASH_REMATCH[1]}"
        local r1="${BASH_REMATCH[2]}"
        
        local re_params_new="${re_params_same_name_prefix}${fname_p}[[:space:]]*\\((.*?)\\)"
        
        for a in "${added[@]}"; do
         if [[ "$a" =~ $re_params_new ]]; then
            local r2="${BASH_REMATCH[1]}"
            if [[ "$r1" != "$r2" ]]; then
              info "- changed return type of $fname_r → MAJOR [$line]"
              major=1
            fi
          fi
        done
      fi

      # Changed parameters
      if [[ "$line" =~ public[[:space:]]+function[[:space:]]+([A-Za-z0-9_]+)[[:space:]]*\((.*?)\) ]]; then
        local fname_p="${BASH_REMATCH[1]}"
        local p1="${BASH_REMATCH[2]}"
        for a in "${added[@]}"; do
          if [[ "$a" =~ public[[:space:]]+function[[:space:]]+${fname_p[[:space:]]}*\((.*?)\) ]]; then
            local p2="${BASH_REMATCH[1]}"
            if [[ "$p1" != "$p2" ]]; then
              info "- changed parameters for $fname_p → MAJOR [$line]"
              major=1
            fi
          fi
        done
      fi

      # Removed public property
      if grep -Eq 'public[[:space:]]+\$[A-Za-z0-9_]+' <<<"$line"; then
        info "- removed public property → MAJOR [$line]"
        major=1
      fi
      # Removed public const
      if grep -Eq 'public[[:space:]]+const[[:space:]]+[A-Za-z0-9_]+' <<<"$line"; then
        info "- removed public constant → MAJOR [$line]"
        major=1
      fi
    done

    # Laravel controller heuristic
    if [[ "$file" == app/Http/Controllers/* && ( $major -eq 1 || $minor -eq 1 ) ]]; then
      info "- controller change detected → Minor [$file]"
      minor=1
    fi
  done

  if ((${#composerFiles[@]} > 0)); then
    info "- composer.json/lock changed → patch"
  fi
  if ((${#views[@]} > 0)); then
    info "- new views → Minor [${views[*]}]"
    minor=1
  fi
  if ((${#migrations[@]} > 0)); then
    info "- new migrations → Minor [${migrations[*]}]"
    minor=1
  fi
  if ((${#configs[@]} > 0)); then
    info "- new configs → Minor [${configs[*]}]"
    minor=1
  fi

  echo

  local nonCodeChanges=0
  if ((${#phpFiles[@]} == 0)) && ( ((${#tests[@]} > 0)) || ((${#docs[@]} > 0)) ); then
    nonCodeChanges=1
  fi

  if ((nonCodeChanges)); then
    info "🧪 Only tests/docs changed → PATCH [${tests[*]}, ${docs[*]}]"
    TYPE="patch"
    return 0
  fi

  if ((major)); then
    info "🧨 Detected MAJOR changes"
    TYPE="major"
  elif ((minor)); then
    info "✨ Detected Minor changes"
    TYPE="minor"
  else
    info "🐛 Only safe changes → PATCH"
    TYPE="patch"
  fi
}

bump_new_version() {
  # Confirm / override type
  printf 'Please confirm release type [major|minor|patch] (default: %s): ' "${TYPE:-patch}"
  read -r answer || true
  if [[ -n "$answer" ]]; then
    TYPE="$answer"
  fi

  local major minor patch
  IFS='.' read -r major minor patch <<<"${OLD_VERSION:-0.0.0}"
  major=${major:-0}
  minor=${minor:-0}
  patch=${patch:-0}

  case "$TYPE" in
    major)
      ((major++))
      minor=0
      patch=0
      ;;
    minor)
      ((minor++))
      patch=0
      ;;
    patch|"")
      ((patch++))
      TYPE="patch"
      ;;
    *)
      warn "Invalid release type: $TYPE"
      return 1
      ;;
  esac

  NEW_VERSION="${major}.${minor}.${patch}"

  if [[ "$OLD_TAG" == v* ]]; then
    NEW_TAG="v${NEW_VERSION}"
  else
    NEW_TAG="$NEW_VERSION"
  fi

  info "Bumping new $TYPE version from $OLD_TAG to $NEW_TAG"
}

create_new_tag() {
  if [[ "$FORCE" -ne 1 ]]; then
    printf 'Are you sure you want to create a new %s %s? [y/N] ' "$TYPE" "$NEW_TAG"
    read -r ans
    case "$ans" in
      y|Y|yes|YES) ;;
      *)
        info "Aborted."
        return 1
        ;;
    esac
  fi

  info "Creating new tag $NEW_TAG and pushing..."
  (
    cd "$GIT_DIR"
    git tag "$NEW_TAG"
    git push
    git push --tags
  ) || {
    warn "Failed to create/push tag $NEW_TAG"
    return 1
  }
}

get_changes() {
  info "Detecting changes for release notes..."
  local log
  local body=""
  local line msg author

  if ! log=$(cd "$GIT_DIR" && git log "$OLD_TAG..HEAD" --pretty='format:%s[####]%an'); then
    CHANGES="## What's Changed"$'\n\n'
    CHANGES+="No commits found since $OLD_TAG"$'\n\n'
    CHANGES+="**Full Changelog**: https://github.com/$GITHUB_REPOSITORY/compare/$OLD_TAG...$NEW_TAG"
    return 0
  fi

  if [[ -z "$log" ]]; then
    CHANGES="## What's Changed"$'\n\n'
    CHANGES+="No commits found"$'\n\n'
    CHANGES+="**Full Changelog**: https://github.com/$GITHUB_REPOSITORY/compare/$OLD_TAG...$NEW_TAG"
    return 0
  fi

  while IFS= read -r line; do
    [[ -z "$line" ]] && continue

    msg="${line%%[####]*}"
    author="${line##*[####]}"

    [[ "$author" == "dependabot[bot]" ]] && author="dependabot"

    body+=$'- '
    body+="$msg"
    body+=" *by $author*"$'\n'
  done <<<"$log"

  CHANGES="## What's Changed"$'\n\n'
  CHANGES+="$body"$'\n'
  CHANGES+="**Full Changelog**: https://github.com/$GITHUB_REPOSITORY/compare/$OLD_TAG...$NEW_TAG"
}

create_release() {
  info "Creating GitHub release $NEW_TAG..."

  local payload resp html_url
  payload=$(jq -n \
    --arg tag "$NEW_TAG" \
    --arg body "$CHANGES" \
    '{tag_name:$tag, name:$tag, body:$body, draft:false, prerelease:false}')

  if ! resp=$(
    curl -sS \
      -H "Authorization: Bearer $GITHUB_TOKEN" \
      -H "Accept: application/vnd.github+json" \
      -d "$payload" \
      "https://api.github.com/repos/$GITHUB_REPOSITORY/releases"
  ); then
    warn "Failed to call GitHub API for release creation"
    return 1
  fi

  html_url=$(jq -r '.html_url // empty' <<<"$resp" || true)

  if [[ -z "$html_url" || "$html_url" == "null" ]]; then
    warn "Failed to create GitHub release; response was:"
    printf '%s\n' "$resp"
    return 1
  fi

  RELEASE_URL="$html_url"
}


########################################
# Main
########################################

main() {
   check_uncommitted_changes
   get_repository
   get_current_version
   detect_release_type
   bump_new_version
   create_new_tag
   get_changes
   create_release

  printf '✅ Created GitHub release: %s\n' "$RELEASE_URL"
}

main
