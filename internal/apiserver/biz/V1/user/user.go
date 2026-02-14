package user

import (
	"aimtp/internal/apiserver/model"
	"aimtp/internal/apiserver/store"
	"aimtp/internal/pkg/contextx"
	"aimtp/internal/pkg/conversion"
	"aimtp/internal/pkg/errno"
	"aimtp/internal/pkg/known"
	"aimtp/internal/pkg/log"
	apiv1 "aimtp/pkg/api/apiserver/v1"
	"aimtp/pkg/authn"
	"aimtp/pkg/authz"
	"aimtp/pkg/store/where"
	"aimtp/pkg/token"
	"context"
	"sync"

	"github.com/jinzhu/copier"
	"golang.org/x/sync/errgroup"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type UserBiz interface {
	Create(ctx context.Context, rq *apiv1.CreateUserRequest) (*apiv1.CreateUserResponse, error)
	Update(ctx context.Context, rq *apiv1.UpdateUserRequest) (*apiv1.UpdateUserResponse, error)
	Delete(ctx context.Context, rq *apiv1.DeleteUserRequest) (*apiv1.DeleteUserResponse, error)
	Get(ctx context.Context, rq *apiv1.GetUserRequest) (*apiv1.GetUserResponse, error)
	List(ctx context.Context, rq *apiv1.ListUserRequest) (*apiv1.ListUserResponse, error)

	UserExpansion
}

type UserExpansion interface {
	Login(ctx context.Context, rq *apiv1.LoginRequest) (*apiv1.LoginResponse, error)
	RefreshToken(ctx context.Context, rq *apiv1.RefreshTokenRequest) (*apiv1.RefreshTokenResponse, error)
	ChangePassword(ctx context.Context, rq *apiv1.ChangePasswordRequest) (*apiv1.ChangePasswordResponse, error)
}

type userBiz struct {
	store store.IStore
	authz *authz.Authz
}

// 确保 userBiz 实现了 UserBiz 接口.
var _ UserBiz = (*userBiz)(nil)

func New(store store.IStore, authz *authz.Authz) *userBiz {
	return &userBiz{
		store: store,
		authz: authz,
	}
}

// Create 实现 UserBiz 接口中的 Create 方法.
func (b *userBiz) Create(ctx context.Context, rq *apiv1.CreateUserRequest) (*apiv1.CreateUserResponse, error) {
	// 注册用户时把明文密码加密操作放在了 hook 中

	var userM model.UserM
	_ = copier.Copy(&userM, rq)

	if err := b.store.User().Create(ctx, &userM); err != nil {
		return nil, err
	}

	// 添加授权角色
	if _, err := b.authz.AddGroupingPolicy(userM.UserID, known.RoleUser); err != nil {
		log.W(ctx).Errorw("Failed to add grouping policy for user", "user", userM.UserID, "role", known.RoleUser)
		return nil, errno.ErrAddRole.WithMessage(err.Error(), "")
	}

	return &apiv1.CreateUserResponse{UserID: userM.UserID}, nil
}

// Update 实现 UserBiz 接口中的 Update 方法.
func (b *userBiz) Update(ctx context.Context, rq *apiv1.UpdateUserRequest) (*apiv1.UpdateUserResponse, error) {
	userM, err := b.store.User().Get(ctx, where.T(ctx))
	if err != nil {
		return nil, err
	}

	if rq.Username != nil {
		userM.Username = *rq.Username
	}

	if rq.Nickname != nil {
		userM.Nickname = *rq.Nickname
	}

	if rq.Email != nil {
		userM.Email = *rq.Email
	}

	if rq.Phone != nil {
		userM.Phone = *rq.Phone
	}

	err = b.store.User().Update(ctx, userM)
	if err != nil {
		return nil, err
	}

	return &apiv1.UpdateUserResponse{}, nil
}

// Delete 实现 UserBiz 接口中的 Delete 方法.
func (b *userBiz) Delete(ctx context.Context, rq *apiv1.DeleteUserRequest) (*apiv1.DeleteUserResponse, error) {
	// 只有 `root` 用户可以删除用户，并且可以删除其他用户
	// 这里不用 where.T()，因为 where.T() 会查询用户自己，而不是查询 rq.UserID 指定的用户
	err := b.store.User().Delete(ctx, where.F("userID", rq.UserID))
	if err != nil {
		return nil, err
	}

	// 删除授权角色
	if _, err := b.authz.RemoveGroupingPolicy(rq.GetUserID(), known.RoleUser); err != nil {
		log.W(ctx).Errorw("Failed to remove grouping policy for user", "user", rq.GetUserID(), "role", known.RoleUser)
		return nil, errno.ErrRemoveRole.WithMessage(err.Error(), "")
	}

	return &apiv1.DeleteUserResponse{}, nil
}

// Get 实现 UserBiz 接口中的 Get 方法.
func (b *userBiz) Get(ctx context.Context, rq *apiv1.GetUserRequest) (*apiv1.GetUserResponse, error) {
	userM, err := b.store.User().Get(ctx, where.F("userID", rq.UserID))
	if err != nil {
		return nil, err
	}

	converted := conversion.UserModelToUserV1(userM)
	return &apiv1.GetUserResponse{
		User: converted,
	}, nil
}

// List 实现 UserBiz 接口中的 List 方法.
func (b *userBiz) List(ctx context.Context, rq *apiv1.ListUserRequest) (*apiv1.ListUserResponse, error) {
	whr := where.P(int(rq.GetOffset()), int(rq.GetLimit()))
	// 如果不是 root 用户，只能查看自己的信息
	if contextx.Username(ctx) != known.AdminUsername {
		whr.T(ctx)
	}

	// 如果是 root 用户，userList 将包含所有用户
	// 如果不是 root 用户，userList 只包含当前用户自己
	count, userList, err := b.store.User().List(ctx, whr)
	if err != nil {
		return nil, err
	}

	var m sync.Map
	eg, ctx := errgroup.WithContext(ctx)

	// 设置最大并发数量为常量 MaxErrGroupConcurrency
	eg.SetLimit(known.MaxErrGroupConcurrency)
	for _, user := range userList {
		// 避免闭包捕获循环变量，如果不这样做可能最后将用到的是切片最后一个 user
		// go 1.22 之后就修复了这个问题
		// u := user
		// 如果有一个 goroutine 返回了 error，其他的 goroutine 会被取消，借助了 ctx
		eg.Go(func() error {
			select {
			case <-ctx.Done():
				return nil
			default:
				// 将 Model 层的 UserM 转为 Protobuf 层的 User
				converted := conversion.UserModelToUserV1(user)
				m.Store(user.UserID, converted)

				return nil
			}
		})
	}

	if err := eg.Wait(); err != nil {
		log.W(ctx).Errorw("Failed to wait all function calls returned", "err", err)
		return nil, err
	}

	users := make([]*apiv1.User, 0, len(userList))
	for _, item := range userList {
		user, _ := m.Load(item.UserID)
		users = append(users, user.(*apiv1.User))
	}

	log.W(ctx).Debugw("Get users from backend storage", "count", len(users))

	return &apiv1.ListUserResponse{
		TotalCount: count,
		Users:      users,
	}, nil
}

// Login 实现 UserBiz 接口中的登陆方法.
func (b *userBiz) Login(ctx context.Context, rq *apiv1.LoginRequest) (*apiv1.LoginResponse, error) {

	userM, err := b.store.User().Get(ctx, where.F("username", rq.Username))
	if err != nil {
		return nil, err
	}

	// 用户登陆需要对加密后的密码进行对比
	if err := authn.Compare(userM.Password, rq.Password); err != nil {
		log.W(ctx).Errorw("Failed to compare password", "err", err)
		return nil, errno.ErrPasswordInvalid
	}

	// 如果匹配成功，说明登录成功，签发 token 并返回
	tk, expiration, err := token.Sign(userM.UserID)
	if err != nil {
		log.W(ctx).Errorw("Failed to sign token", "err", err)
		return nil, errno.ErrSignToken
	}

	return &apiv1.LoginResponse{Token: tk, ExpireAt: timestamppb.New(expiration)}, nil
}

// RefreshToken 用于刷新用户的身份验证令牌.
// 当用户的令牌即将过期时，可以调用此方法生成一个新的令牌.
func (b *userBiz) RefreshToken(ctx context.Context, rq *apiv1.RefreshTokenRequest) (*apiv1.RefreshTokenResponse, error) {
	tk, expiration, err := token.Sign(contextx.UserID(ctx))
	if err != nil {
		log.W(ctx).Errorw("Failed to sign token", "err", err)
		return nil, errno.ErrSignToken
	}

	return &apiv1.RefreshTokenResponse{Token: tk, ExpireAt: timestamppb.New(expiration)}, nil
}

// ChangePassword 实现 UserBiz 接口中的修改密码方法.
// 用户需要提供旧密码以验证身份，然后才能修改为新密码.
func (b *userBiz) ChangePassword(ctx context.Context, rq *apiv1.ChangePasswordRequest) (*apiv1.ChangePasswordResponse, error) {
	userM, err := b.store.User().Get(ctx, where.T(ctx))
	if err != nil {
		return nil, err
	}

	if err := authn.Compare(userM.Password, rq.GetOldPassword()); err != nil {
		log.W(ctx).Errorw("Failed to compare password", "err", err)
		return nil, errno.ErrPasswordInvalid
	}

	userM.Password, _ = authn.Encrypt(rq.GetNewPassword())
	if err := b.store.User().Update(ctx, userM); err != nil {
		return nil, err
	}

	return &apiv1.ChangePasswordResponse{}, nil
}
