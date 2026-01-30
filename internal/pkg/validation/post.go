package validation

import (
	"context"
	"aimtp/internal/pkg/errno"
	apiv1 "aimtp/pkg/api/apiserver/v1"
	genericvalidation "aimtp/pkg/validation"
)

func (v *Validator) ValidatePostRules() genericvalidation.Rules {
	return genericvalidation.Rules{
		"Title": func(value any) error {
			if value.(string) == "" {
				return errno.ErrInvalidArgument.WithMessage("title cannot be empty")
			}
			return nil
		},
		"PostID": func(value any) error {
			if value.(string) == "" {
				return errno.ErrInvalidArgument.WithMessage("postID cannot be empty")
			}
			return nil
		},
		"PostIDs": func(value any) error {
			if len(value.([]string)) == 0 {
				return errno.ErrInvalidArgument.WithMessage("postIDs cannot be empty")
			}
			for _, postID := range value.([]string) {
				if postID == "" {
					return errno.ErrInvalidArgument.WithMessage("postID cannot be empty")
				}
			}
			return nil
		},
		"Offset": func(value any) error {
			if value.(int64) < 0 {
				return errno.ErrInvalidArgument.WithMessage("limit must be greater than 0")
			}
			return nil
		},
		"Limit": func(value any) error {
			if value.(int64) <= 0 {
				return errno.ErrInvalidArgument.WithMessage("offset cannot be negative")
			}
			return nil
		},
	}
}

func (v *Validator) ValidateCreatePostRequest(ctx context.Context, rq *apiv1.CreatePostRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidatePostRules())
}

func (v *Validator) ValidateUpdatePostRequest(ctx context.Context, rq *apiv1.UpdatePostRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidatePostRules())
}

func (v *Validator) ValidateDeletePostRequest(ctx context.Context, rq *apiv1.DeletePostRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidatePostRules())
}

func (v *Validator) ValidateGetPostRequest(ctx context.Context, rq *apiv1.GetPostRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidatePostRules())
}

func (v *Validator) ValidateListPostRequest(ctx context.Context, rq *apiv1.ListPostRequest) error {
	return genericvalidation.ValidateAllFields(rq, v.ValidatePostRules())
}
