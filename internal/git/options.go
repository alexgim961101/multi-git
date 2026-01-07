package git

import (
	"io"
	"time"
)

// CloneOptions represents options for cloning a repository
type CloneOptions struct {
	Depth    int       // Shallow clone depth (0 = full clone)
	Branch   string    // 특정 브랜치만 클론
	Progress io.Writer // 진행 상황 출력 (nil이면 출력 안 함)
}

// CheckoutOptions represents options for checking out a branch
type CheckoutOptions struct {
	Branch     string // 체크아웃할 브랜치 이름
	Create     bool   // 브랜치가 없으면 생성
	Force      bool   // 로컬 변경사항 무시하고 강제 체크아웃
	FetchFirst bool   // 체크아웃 전 fetch 수행
}

// TagOptions represents options for tag operations
type TagOptions struct {
	Name      string // 태그 이름
	Message   string // 태그 메시지 (annotated tag용)
	Annotated bool   // annotated tag (true) vs lightweight tag (false)
	Force     bool   // 기존 태그 덮어쓰기
	Push      bool   // 원격에 푸시
}

// PushOptions represents options for pushing to remote
type PushOptions struct {
	Branch  string        // 푸시할 브랜치 이름
	Remote  string        // 원격 이름 (기본: origin)
	Force   bool          // 강제 푸시
	DryRun  bool          // 시뮬레이션만 (실제 푸시 안 함)
	Timeout time.Duration // 타임아웃 (0 = 기본값)
}

// AuthOptions represents authentication options
type AuthOptions struct {
	Username string // 사용자 이름 (HTTPS용)
	Password string // 비밀번호 또는 토큰 (HTTPS용)
	// SSH 키는 시스템 기본값 사용
}
