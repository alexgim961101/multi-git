#!/bin/bash
#
# Multi-Git 제거 스크립트
# 사용법: ./scripts/uninstall.sh [OPTIONS]
#
# OPTIONS:
#   --prefix=PATH    설치 경로 지정 (기본: /usr/local/bin)
#   --user           사용자 홈 디렉토리에서 제거 (~/.local/bin)
#   --all            설정 파일도 함께 제거
#   --help           도움말 출력
#

set -e

# 색상 정의
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 기본 설정
INSTALL_DIR="/usr/local/bin"
BINARY_NAME="multi-git"
REMOVE_CONFIG=false

# 함수 정의
print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1" >&2
}

print_warning() {
    echo -e "${YELLOW}!${NC} $1"
}

print_help() {
    cat << EOF
Multi-Git 제거 스크립트

사용법: ./scripts/uninstall.sh [OPTIONS]

OPTIONS:
  --prefix=PATH    설치 경로 지정 (기본: /usr/local/bin)
  --user           사용자 홈 디렉토리에서 제거 (~/.local/bin)
  --all            설정 파일도 함께 제거
  --help           도움말 출력

예시:
  ./scripts/uninstall.sh                 # /usr/local/bin에서 제거
  ./scripts/uninstall.sh --user          # ~/.local/bin에서 제거
  ./scripts/uninstall.sh --user --all    # 바이너리와 설정 파일 모두 제거
EOF
}

# 인자 파싱
for arg in "$@"; do
    case $arg in
        --prefix=*)
            INSTALL_DIR="${arg#*=}"
            shift
            ;;
        --user)
            INSTALL_DIR="$HOME/.local/bin"
            shift
            ;;
        --all)
            REMOVE_CONFIG=true
            shift
            ;;
        --help)
            print_help
            exit 0
            ;;
        *)
            print_error "알 수 없는 옵션: $arg"
            print_help
            exit 1
            ;;
    esac
done

echo "========================================="
echo "  Multi-Git 제거 스크립트"
echo "========================================="
echo ""

# 바이너리 제거
BINARY_PATH="$INSTALL_DIR/$BINARY_NAME"

if [ -f "$BINARY_PATH" ]; then
    echo "바이너리 제거 중..."
    if rm "$BINARY_PATH" 2>/dev/null; then
        print_success "바이너리 제거 완료: $BINARY_PATH"
    elif sudo rm "$BINARY_PATH"; then
        print_success "바이너리 제거 완료: $BINARY_PATH (sudo)"
    else
        print_error "바이너리 제거 실패: $BINARY_PATH"
        exit 1
    fi
else
    print_warning "바이너리를 찾을 수 없습니다: $BINARY_PATH"
fi

# 설정 파일 제거
if [ "$REMOVE_CONFIG" = true ]; then
    CONFIG_DIR="$HOME/.multi-git"
    
    if [ -d "$CONFIG_DIR" ]; then
        echo ""
        echo "설정 파일 제거 중..."
        rm -rf "$CONFIG_DIR"
        print_success "설정 디렉토리 제거 완료: $CONFIG_DIR"
    else
        print_warning "설정 디렉토리를 찾을 수 없습니다: $CONFIG_DIR"
    fi
fi

# 완료 메시지
echo ""
echo "========================================="
echo "  제거 완료!"
echo "========================================="
echo ""

if [ "$REMOVE_CONFIG" = false ]; then
    echo "설정 파일은 유지됩니다: ~/.multi-git/"
    echo "설정 파일도 제거하려면 --all 옵션을 사용하세요."
    echo ""
fi

