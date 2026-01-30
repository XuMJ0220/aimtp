package conversion

import (
	"aimtp/internal/apiserver/model"
	apiv1 "aimtp/pkg/api/apiserver/v1"
	"aimtp/pkg/core"
)

// UserModelToUserV1 将模型层的 UserM 转换为 Protobuf 层的 User
func UserModelToUserV1(userModel *model.UserM) *apiv1.User {
	var protoBuf apiv1.User
	_ = core.CopyWithConverters(&protoBuf, userModel)
	return &protoBuf
}

// UserV1ToUserModel 将 Protobuf 层的 User 转换为模型层的 UserM
func UserV1ToUserModel(protoUser *apiv1.User) *model.UserM {
	var userModel model.UserM
	_ = core.CopyWithConverters(&userModel, protoUser)
	return &userModel
}
