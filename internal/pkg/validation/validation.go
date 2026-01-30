package validation

import (
	"aimtp/internal/apiserver/store"
	"aimtp/internal/pkg/errno"
	"regexp"
)

// Validator 是验证逻辑的实现结构体.
type Validator struct {
	// 有些复杂的验证逻辑，可能需要直接查询数据库
	// 这里只是一个举例，如果验证时，有其他依赖的客户端/服务/资源等，
	// 都可以一并注入进来
	store store.IStore
}

// New 创建一个新的 Validator 实例.
func New(store store.IStore) *Validator {
	return &Validator{store: store}
}

func isValiPassword(password string) error {
	if len(password) < 8 || len(password) > 32 {
		return errno.ErrInvalidArgument.WithMessage("password must be between 8 and 32 characters long")
	}

	// 2. 必须包含字母
	hasLetter := regexp.MustCompile(`[A-Za-z]`).MatchString(password)
	// 3. 必须包含数字
	hasNumber := regexp.MustCompile(`\d`).MatchString(password)
	// 4. 必须包含特殊字符
	hasSpecial := regexp.MustCompile(`[\W_]`).MatchString(password)

	if !hasLetter || !hasNumber || !hasSpecial {
		return errno.ErrInvalidArgument.WithMessage("password must contain at least one letter, one number, and one special character")
	}

	return nil
}

func isValiName(name string) error {
	if len(name) < 4 || len(name) > 20 {
		return errno.ErrInvalidArgument.WithMessage("name must be between 6 and 20 characters long")
	}

	// 只能由字母,数字,下划线,减号组成
	formatPattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !formatPattern.MatchString(name) {
		return errno.ErrInvalidArgument.WithMessage("name can only contain letters, numbers, underscores, and hyphens")
	}

	return nil
}
