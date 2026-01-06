# Multi-Git

여러 Git 저장소를 효율적으로 관리하는 CLI 도구입니다. DevOps 직군의 개발자들이 여러 저장소에 대한 반복적인 작업을 자동화할 수 있도록 도와줍니다.

## 📋 목차

- [주요 기능](#주요-기능)
- [설치 방법](#설치-방법)
- [빠른 시작](#빠른-시작)
- [설정 파일](#설정-파일)
- [사용법](#사용법)
- [예제](#예제)
- [기여하기](#기여하기)
- [라이선스](#라이선스)

## ✨ 주요 기능

- **저장소 일괄 클론**: 여러 Git 저장소를 한 번에 클론
- **브랜치 일괄 변경**: 관리되는 모든 저장소의 동일한 브랜치로 한 번에 체크아웃
- **태그 동시 관리**: 여러 저장소의 특정 브랜치에 태그를 동시에 생성/푸시
- **강제 Push**: 릴리스 배포 시 브랜치 충돌 해결을 위한 강제 push 지원

## 🚀 설치 방법

### 요구사항

- Go 1.24 이상
- Git 2.0 이상

### 소스에서 빌드

```bash
# 저장소 클론
git clone https://github.com/lotto/multi-git.git
cd multi-git

# 빌드
go build -o multi-git cmd/multi-git/main.go

# 설치 (선택사항)
sudo mv multi-git /usr/local/bin/
```

### 바이너리 다운로드 (향후 제공 예정)

릴리스 페이지에서 운영체제에 맞는 바이너리를 다운로드할 수 있습니다.

## 🏃 빠른 시작

1. **설정 파일 생성**

```bash
mkdir -p ~/.multi-git
cat > ~/.multi-git/config.yaml << EOF
config:
  base_dir: ~/repositories
  default_remote: origin
  parallel_workers: 3

repositories:
  - name: backend-service
    url: https://github.com/org/backend-service.git
  
  - name: frontend-app
    url: https://github.com/org/frontend-app.git
EOF
```

2. **저장소 클론**

```bash
multi-git clone
```

3. **브랜치 변경**

```bash
multi-git checkout release/v1.0.0
```

## ⚙️ 설정 파일

설정 파일은 기본적으로 `~/.multi-git/config.yaml`에 위치합니다. `--config` 플래그로 다른 경로를 지정할 수 있습니다.

### 설정 파일 구조

```yaml
config:
  base_dir: ~/repositories      # 저장소 클론 기본 디렉토리
  default_remote: origin         # 기본 원격 이름
  parallel_workers: 3            # 병렬 작업 수

repositories:
  - name: backend-service        # 저장소 이름
    url: https://github.com/org/backend-service.git  # 저장소 URL
    path: backend                # 선택적 경로 오버라이드
  
  - name: frontend-app
    url: https://github.com/org/frontend-app.git
    # path가 없으면 name을 사용
```

### 저장소 URL 형식

- HTTPS: `https://github.com/org/repo.git`
- SSH: `git@github.com:org/repo.git`

## 📖 사용법

### `clone` - 저장소 클론

여러 저장소를 한 번에 클론합니다.

```bash
multi-git clone [flags]
```

**Flags:**
- `--config, -c`: 설정 파일 경로 (기본값: `~/.multi-git/config.yaml`)
- `--skip-existing`: 이미 존재하는 저장소 스킵 (기본값: `true`)
- `--parallel, -p`: 병렬 클론 개수 (기본값: `3`)
- `--depth`: Shallow clone depth (선택사항)

**예제:**
```bash
# 기본 클론
multi-git clone

# 병렬 클론 수 지정
multi-git clone --parallel 5

# 이미 존재하는 저장소도 다시 클론
multi-git clone --skip-existing=false
```

### `checkout` - 브랜치 일괄 변경

관리되는 모든 저장소의 동일한 브랜치로 한 번에 체크아웃합니다.

```bash
multi-git checkout <branch-name> [flags]
```

**Flags:**
- `--create, -c`: 브랜치가 없으면 생성
- `--force, -f`: 로컬 변경사항 무시하고 강제 체크아웃
- `--fetch`: 체크아웃 전 fetch 수행

**예제:**
```bash
# 브랜치 변경
multi-git checkout release/v1.0.0

# 브랜치가 없으면 생성
multi-git checkout feature/new-feature --create

# fetch 후 체크아웃
multi-git checkout release/v1.0.0 --fetch
```

### `tag` - 태그 관리

여러 저장소의 특정 브랜치에 태그를 동시에 생성/푸시합니다.

```bash
multi-git tag --branch <branch> --name <tag-name> [flags]
```

**Flags:**
- `--branch, -b`: 태그를 생성할 브랜치 이름 (필수)
- `--name, -n`: 태그 이름 (필수)
- `--message, -m`: 태그 메시지
- `--push, -p`: 태그를 원격에 푸시
- `--force, -f`: 기존 태그 덮어쓰기
- `--delete, -d`: 태그 삭제

**예제:**
```bash
# 태그 생성
multi-git tag --branch release/v1.0.0 --name v1.0.0

# 태그 생성 및 푸시
multi-git tag --branch release/v1.0.0 --name v1.0.0 --push --message "Release v1.0.0"

# 태그 삭제
multi-git tag --name v1.0.0 --delete --push
```

### `push` - 강제 Push

여러 저장소의 특정 브랜치에 강제 push를 수행합니다.

```bash
multi-git push --branch <branch> --force [flags]
```

**Flags:**
- `--branch, -b`: 푸시할 브랜치 이름 (필수)
- `--force, -f`: 강제 push (필수)
- `--remote, -r`: 원격 이름 (기본값: `origin`)
- `--dry-run`: 실제 푸시 없이 시뮬레이션
- `--yes, -y`: 확인 프롬프트 스킵

**예제:**
```bash
# 강제 push (확인 프롬프트 표시)
multi-git push --branch release/v1.0.0 --force

# 확인 프롬프트 스킵
multi-git push --branch release/v1.0.0 --force --yes

# 시뮬레이션만 실행
multi-git push --branch release/v1.0.0 --force --dry-run
```

## 💡 예제

### 시나리오 1: 릴리스 준비

```bash
# 1. 모든 저장소를 release 브랜치로 변경
multi-git checkout release/v1.0.0 --fetch

# 2. 릴리스 태그 생성 및 푸시
multi-git tag --branch release/v1.0.0 --name v1.0.0 --push --message "Release v1.0.0"
```

### 시나리오 2: 배포 후 충돌 해결

```bash
# 브랜치 충돌 해결을 위한 강제 push
multi-git push --branch release/v1.0.0 --force --yes
```

### 시나리오 3: 새 프로젝트 설정

```bash
# 1. 모든 저장소 클론
multi-git clone

# 2. 개발 브랜치로 변경
multi-git checkout develop --fetch
```

## 🛠️ 개발

### 프로젝트 구조

```
multi-git/
├── cmd/
│   └── multi-git/          # CLI 진입점
├── internal/
│   ├── commands/           # 명령어 구현
│   ├── config/             # 설정 관리
│   ├── repository/         # 저장소 관리
│   └── git/                # Git 작업
├── pkg/
│   └── errors/             # 에러 타입
└── docs/                    # 문서
```

### 빌드

```bash
go build -o multi-git cmd/multi-git/main.go
```

### 테스트

```bash
go test ./...
```

## 🤝 기여하기

기여를 환영합니다! 이슈를 생성하거나 Pull Request를 제출해주세요.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📝 라이선스

이 프로젝트는 MIT 라이선스를 따릅니다.

## 📚 관련 문서

- [PRD](./docs/PRD.md) - 제품 요구사항 문서
- [Tech Spec](./docs/TECH_SPEC.md) - 기술 명세서
- [Implementation Plan](./docs/IMPLEMENTATION_PLAN.md) - 구현 계획서

## 🐛 문제 리포트

버그를 발견하셨나요? [이슈를 생성](https://github.com/lotto/multi-git/issues)해주세요.

