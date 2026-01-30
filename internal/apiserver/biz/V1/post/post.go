package post

import (
	"context"
	"aimtp/internal/apiserver/model"
	"aimtp/internal/apiserver/store"
	"aimtp/internal/pkg/conversion"
	apiv1 "aimtp/pkg/api/apiserver/v1"
	"aimtp/pkg/store/where"

	"github.com/jinzhu/copier"
)

type PostBiz interface {
	Create(ctx context.Context, rq *apiv1.CreatePostRequest) (*apiv1.CreatePostResponse, error)
	Update(ctx context.Context, rq *apiv1.UpdatePostRequest) (*apiv1.UpdatePostResponse, error)
	Delete(ctx context.Context, rq *apiv1.DeletePostRequest) (*apiv1.DeletePostResponse, error)
	Get(ctx context.Context, rq *apiv1.GetPostRequest) (*apiv1.GetPostResponse, error)
	List(ctx context.Context, rq *apiv1.ListPostRequest) (*apiv1.ListPostResponse, error)

	PostExpansion
}

type PostExpansion interface {
}

type postBiz struct {
	store store.IStore
}

// 确保 postBiz 实现了 PostBiz 接口.
var _ PostBiz = (*postBiz)(nil)

func New(store store.IStore) *postBiz {
	return &postBiz{
		store: store,
	}
}

// Create 实现 PostBiz 接口中的 Create 方法.
func (b *postBiz) Create(ctx context.Context, rq *apiv1.CreatePostRequest) (*apiv1.CreatePostResponse, error) {
	var postM model.PostM
	_ = copier.Copy(&postM, rq)

	if err := b.store.Post().Create(ctx, &postM); err != nil {
		return nil, err
	}

	return &apiv1.CreatePostResponse{
		PostID: postM.PostID,
	}, nil
}

// Update 实现 PostBiz 接口中的 Update 方法.
func (b *postBiz) Update(ctx context.Context, rq *apiv1.UpdatePostRequest) (*apiv1.UpdatePostResponse, error) {
	postM, err := b.store.Post().Get(ctx, where.F("postID", rq.GetPostID()))
	if err != nil {
		return nil, err
	}

	if rq.Title != nil {
		postM.Title = rq.GetTitle()
	}

	if rq.Content != nil {
		postM.Content = rq.GetContent()
	}

	if err := b.store.Post().Update(ctx, postM); err != nil {
		return nil, err
	}

	return &apiv1.UpdatePostResponse{}, nil
}

// Delete 实现 PostBiz 接口中的 Delete 方法.
func (b *postBiz) Delete(ctx context.Context, rq *apiv1.DeletePostRequest) (*apiv1.DeletePostResponse, error) {
	if err := b.store.Post().Delete(ctx, where.F("postID", rq.GetPostIDs())); err != nil {
		return nil, err
	}

	return &apiv1.DeletePostResponse{}, nil
}

// Get 实现 PostBiz 接口中的 Get 方法.
func (b *postBiz) Get(ctx context.Context, rq *apiv1.GetPostRequest) (*apiv1.GetPostResponse, error) {
	postM, err := b.store.Post().Get(ctx, where.F("postID", rq.GetPostID()))
	if err != nil {
		return nil, err
	}

	return &apiv1.GetPostResponse{
		Post: conversion.PostModelToPostV1(postM),
	}, nil
}

// List 实现 PostBiz 接口中的 List 方法.
func (b *postBiz) List(ctx context.Context, rq *apiv1.ListPostRequest) (*apiv1.ListPostResponse, error) {
	whr := where.P(int(rq.GetOffset()), int(rq.GetLimit()))
	if rq.Title != nil {
		whr.F("title", rq.GetTitle())
	}

	count, postList, err := b.store.Post().List(ctx, whr)
	if err != nil {
		return nil, err
	}

	posts := make([]*apiv1.Post, 0, len(postList))
	for _, postM := range postList {
		convented := conversion.PostModelToPostV1(postM)
		posts = append(posts, convented)
	}

	return &apiv1.ListPostResponse{
		TotalCount: count,
		Posts:      posts,
	}, nil
}
