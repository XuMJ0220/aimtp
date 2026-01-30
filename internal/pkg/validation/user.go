package validation

import (
	"context"
	"aimtp/internal/pkg/errno"
	apiv1 "aimtp/pkg/api/apiserver/v1"
	genericvalidation "aimtp/pkg/validation"
	"regexp"
)

func (v *Validator) ValidateUserRules() genericvalidation.Rules {
	var (
		validatePassword genericvalidation.ValidatorFunc
		validateName     genericvalidation.ValidatorFunc
	)

	validatePassword = func(value any) error {
		return isValiPassword(value.(string))
	}

	validateName = func(value any) error {
		return isValiName(value.(string))
	}
	return genericvalidation.Rules{

		"Password":    validatePassword,
		"OldPassword": validatePassword,
		"NewPassword": validatePassword,
		"Username":    validateName,
		"Nickname":    validateName,
		"Email": func(value any) error {
			formatPattern := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
			if !formatPattern.MatchString(value.(string)) {
				return errno.ErrInvalidArgument.WithMessage("invalid email format")
			}
			return nil
		},
		"Phone": func(value any) error {
			formatPattern := regexp.MustCompile(`^1[3-9]\d{9}$`)
			if !formatPattern.MatchString(value.(string)) {
				return errno.ErrInvalidArgument.WithMessage("phone must be a valid 11-digit mobile number")
			}
			return nil
		},
		"Offset": func(value any) error {
			if value.(int64) < 0 {
				return errno.ErrInvalidArgument.WithMessage("offset must be greater than 0")
			}
			return nil
		},
		"Limit": func(value any) error {
			if value.(int64) <= 0 {
				return errno.ErrInvalidArgument.WithMessage("limit cannot be negative")
			}
			return nil
		},
	}
}

func (v *Validator) ValidateCreateUserRequest(ctx context.Context, rq *apiv1.CreateUserRequest) error {
	if rq.Password != rq.Repassword {
		return errno.ErrInvalidArgument.WithMessage("confirmation password must match the password ")
	}
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

func (v *Validator) ValidateUpdateUserRequest(ctx context.Context, rq *apiv1.UpdateUserRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

func (v *Validator) ValidateDeleteUserRequest(ctx context.Context, rq *apiv1.DeleteUserRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

func (v *Validator) ValidateGetUserRequest(ctx context.Context, rq *apiv1.GetUserRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

func (v *Validator) ValidateListUserRequest(ctx context.Context, rq *apiv1.ListUserRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

func (v *Validator) ValidateLoginRequest(ctx context.Context, rq *apiv1.LoginRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}

func (v *Validator) ValidateChangePasswordRequest(ctx context.Context, rq *apiv1.ChangePasswordRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidateUserRules())
}
