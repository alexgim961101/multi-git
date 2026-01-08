#!/bin/bash
#
# Multi-Git 설치 스크립트
# 사용법: ./scripts/install.sh [OPTIONS]
#
# OPTIONS:
#   --prefix=PATH    설치 경로 지정 (기본: /usr/local/bin)
#   --user           사용자 홈 디렉토리에 설치 (~/.local/bin)
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
PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

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
Multi-Git 설치 스크립트

사용법: ./scripts/install.sh [OPTIONS]

OPTIONS:
  --prefix=PATH    설치 경로 지정 (기본: /usr/local/bin)
  --user           사용자 홈 디렉토리에 설치 (~/.local/bin)
  --help           도움말 출력

예시:
  ./scripts/install.sh                    # /usr/local/bin에 설치 (sudo 필요)
  ./scripts/install.sh --user             # ~/.local/bin에 설치
  ./scripts/install.sh --prefix=/opt/bin  # /opt/bin에 설치
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
echo "  Multi-Git 설치 스크립트"
echo "========================================="
echo ""

# Go 설치 확인
echo "Go 설치 확인..."
if ! command -v go &> /dev/null; then
    print_error "Go가 설치되어 있지 않습니다."
    echo "Go 1.24 이상을 설치해주세요: https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
print_success "Go 버전: $GO_VERSION"

# 프로젝트 디렉토리 확인
echo ""
echo "프로젝트 디렉토리 확인..."
if [ ! -f "$PROJECT_DIR/go.mod" ]; then
    print_error "프로젝트 디렉토리를 찾을 수 없습니다: $PROJECT_DIR"
    exit 1
fi
print_success "프로젝트 디렉토리: $PROJECT_DIR"

# 빌드
echo ""
echo "빌드 중..."
cd "$PROJECT_DIR"

if go build -o "$BINARY_NAME" ./cmd/multi-git; then
    print_success "빌드 완료: $PROJECT_DIR/$BINARY_NAME"
else
    print_error "빌드 실패"
    exit 1
fi

# 설치 디렉토리 생성
echo ""
echo "설치 디렉토리 확인..."
if [ ! -d "$INSTALL_DIR" ]; then
    echo "디렉토리 생성: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR" || {
        print_error "디렉토리 생성 실패. sudo로 다시 시도합니다..."
        sudo mkdir -p "$INSTALL_DIR"
    }
fi
print_success "설치 디렉토리: $INSTALL_DIR"

# 바이너리 복사
echo ""
echo "바이너리 설치 중..."
if cp "$BINARY_NAME" "$INSTALL_DIR/" 2>/dev/null; then
    print_success "설치 완료: $INSTALL_DIR/$BINARY_NAME"
elif sudo cp "$BINARY_NAME" "$INSTALL_DIR/"; then
    print_success "설치 완료: $INSTALL_DIR/$BINARY_NAME (sudo)"
else
    print_error "설치 실패"
    exit 1
fi

# 실행 권한 설정
chmod +x "$INSTALL_DIR/$BINARY_NAME" 2>/dev/null || sudo chmod +x "$INSTALL_DIR/$BINARY_NAME"

# 빌드 결과물 정리
rm -f "$PROJECT_DIR/$BINARY_NAME"

# PATH 확인
echo ""
if echo "$PATH" | grep -q "$INSTALL_DIR"; then
    print_success "PATH에 $INSTALL_DIR이 포함되어 있습니다."
else
    print_warning "PATH에 $INSTALL_DIR이 포함되어 있지 않습니다."
    echo ""
    echo "다음 명령어를 실행하거나 ~/.bashrc 또는 ~/.zshrc에 추가하세요:"
    echo ""
    echo "  export PATH=\"\$PATH:$INSTALL_DIR\""
    echo ""
fi

# 설정 파일 예시 생성
CONFIG_DIR="$HOME/.multi-git"
CONFIG_FILE="$CONFIG_DIR/config.yaml"

if [ ! -f "$CONFIG_FILE" ]; then
    echo ""
    echo "설정 파일 예시 생성 중..."
    mkdir -p "$CONFIG_DIR"
    cat > "$CONFIG_FILE" << 'EOF'
# Multi-Git 설정 파일
# 경로: ~/.multi-git/config.yaml

config:
  # 저장소가 클론될 기본 디렉토리
  base_dir: ~/repos
  
  # 기본 원격 저장소 이름
  default_remote: origin
  
  # 병렬 작업 수
  parallel_workers: 3

# 관리할 저장소 목록
repositories:
  # 예시:
  # - name: my-repo
  #   url: https://github.com/username/my-repo.git
  #   path: custom/path  # 선택사항: base_dir 기준 상대 경로
EOF
    print_success "설정 파일 생성: $CONFIG_FILE"
fi

# 완료 메시지
echo ""
echo "========================================="
echo "  설치 완료!"
echo "========================================="
echo ""
echo "사용 방법:"
echo "  $BINARY_NAME --help          # 도움말"
echo "  $BINARY_NAME clone           # 저장소 클론"
echo "  $BINARY_NAME checkout <branch>  # 브랜치 변경"
echo ""
echo "설정 파일: $CONFIG_FILE"
echo ""

