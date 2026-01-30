package conversion

import (
	"aimtp/internal/apiserver/model"
	apiv1 "aimtp/pkg/api/apiserver/v1"
	"aimtp/pkg/core"
)

// PostModelToPostV1 将模型层的 PostM 转换为 Protobuf 层的 Post
func PostModelToPostV1(postModel *model.PostM) *apiv1.Post {
	var protoBuf apiv1.Post
	_ = core.CopyWithConverters(&protoBuf, postModel)
	return &protoBuf
}

// PostV1ToPostModel 将 Protobuf 层的 Post 转换为模型层的 PostM
func PostV1ToPostModel(protoPost *apiv1.Post) *model.PostM {
	var postModel model.PostM
	_ = core.CopyWithConverters(&postModel, protoPost)
	return &postModel
}
