#!/bin/bash

set -e

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 检查是否在 git 仓库中
if ! git rev-parse --git-dir > /dev/null 2>&1; then
    log_error "Not in a git repository"
    exit 1
fi

# 检查是否有未提交的更改
log_info "Checking for uncommitted changes..."
if ! git diff-index --quiet HEAD --; then
    log_error "There are uncommitted changes. Please commit or stash them first."
    git status --short
    exit 1
fi

# 检查是否有未追踪的文件
if [ -n "$(git ls-files --others --exclude-standard)" ]; then
    log_error "There are untracked files. Please add or ignore them first."
    git ls-files --others --exclude-standard
    exit 1
fi

# 检查是否有未推送的提交
log_info "Checking for unpushed commits..."
LOCAL=$(git rev-parse @)
REMOTE=$(git rev-parse @{u} 2>/dev/null || echo "")

if [ -z "$REMOTE" ]; then
    log_warning "No remote tracking branch found. Skipping remote check."
elif [ "$LOCAL" != "$REMOTE" ]; then
    log_error "There are unpushed commits. Please push them first."
    git log --oneline @{u}..@ | head -5
    exit 1
fi

log_success "Working directory is clean and up to date"

# 获取所有子模块目录（包含 go.mod 的目录，排除根目录）
MODULES=()
for dir in */; do
    if [ -f "${dir}go.mod" ]; then
        module_name="${dir%/}"
        MODULES+=("$module_name")
    fi
done

if [ ${#MODULES[@]} -eq 0 ]; then
    log_warning "No submodules found"
    exit 0
fi

log_info "Found ${#MODULES[@]} modules: ${MODULES[*]}"
echo ""

# 递增版本号函数
increment_version() {
    local version=$1
    # 移除 v 前缀
    version=${version#v}
    # 分割版本号
    IFS='.' read -r major minor patch <<< "$version"
    # 递增 patch 版本
    patch=$((patch + 1))
    echo "v${major}.${minor}.${patch}"
}

# 处理每个模块
for module in "${MODULES[@]}"; do
    log_info "================================================"
    log_info "Processing module: $module"
    log_info "================================================"

    # 获取模块的标签前缀
    tag_prefix="${module}/"

    # 获取该模块的最新版本标签
    latest_tag=$(git tag -l "${tag_prefix}v*" | sort -V | tail -n 1)

    if [ -z "$latest_tag" ]; then
        # 没有发布过版本
        new_version="v0.0.1"
        log_info "No previous version found. Will create initial version: $new_version"
        should_release=true
    else
        # 已有版本，检查是否有变化
        current_version="${latest_tag#${tag_prefix}}"
        log_info "Latest version: $current_version (tag: $latest_tag)"

        # 获取该标签的 commit
        tag_commit=$(git rev-list -n 1 "$latest_tag")

        # 检查从该标签到现在，该目录是否有变化
        changes=$(git diff --name-only "$tag_commit" HEAD -- "$module/")

        if [ -z "$changes" ]; then
            log_success "No changes detected since $current_version. Skipping release."
            should_release=false
        else
            log_info "Changes detected since $current_version:"
            echo "$changes" | sed 's/^/  /'
            new_version=$(increment_version "$current_version")
            log_info "Will create new version: $new_version"
            should_release=true
        fi
    fi

    # 如果需要发布
    if [ "$should_release" = true ]; then
        new_tag="${tag_prefix}${new_version}"

        log_info "Creating tag: $new_tag"
        git tag -a "$new_tag" -m "Release $module $new_version"

        log_info "Pushing tag: $new_tag"
        git push origin "$new_tag"

        log_success "Successfully released $module $new_version"
    fi

    echo ""
done

log_success "================================================"
log_success "Release process completed!"
log_success "================================================"

