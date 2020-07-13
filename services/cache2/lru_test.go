// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package cache2

import (
	"fmt"
	"testing"
	"time"

	"github.com/nonce/town-server/model"
	"github.com/nonce/town-server/services/cache/lru"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLRU(t *testing.T) {
	l := NewLRU(&LRUOptions{
		Size:                   128,
		DefaultExpiry:          0,
		InvalidateClusterEvent: "",
	})

	for i := 0; i < 256; i++ {
		err := l.Set(fmt.Sprintf("%d", i), i)
		require.Nil(t, err)
	}
	size, err := l.Len()
	require.Nil(t, err)
	require.Equalf(t, size, 128, "bad len: %v", size)

	keys, err := l.Keys()
	require.Nil(t, err)
	for i, k := range keys {
		var v int
		err = l.Get(k, &v)
		require.Nil(t, err, "bad key: %v", k)
		require.Equalf(t, fmt.Sprintf("%d", v), k, "bad key: %v", k)
		require.Equalf(t, i+128, v, "bad value: %v", k)
	}
	for i := 0; i < 128; i++ {
		var v int
		err = l.Get(fmt.Sprintf("%d", i), &v)
		require.Equal(t, ErrKeyNotFound, err, "should be evicted %v: %v", i, err)
	}
	for i := 128; i < 256; i++ {
		var v int
		err = l.Get(fmt.Sprintf("%d", i), &v)
		require.Nil(t, err, "should not be evicted %v: %v", i, err)
	}
	for i := 128; i < 192; i++ {
		l.Remove(fmt.Sprintf("%d", i))
		var v int
		err = l.Get(fmt.Sprintf("%d", i), &v)
		require.Equal(t, ErrKeyNotFound, err, "should be deleted %v: %v", i, err)
	}

	var v int
	err = l.Get("192", &v) // expect 192 to be last key in l.Keys()
	require.Nil(t, err, "should exist")
	require.Equalf(t, 192, v, "bad value: %v", v)

	keys, err = l.Keys()
	require.Nil(t, err)
	for i, k := range keys {
		require.Falsef(t, i < 63 && k != fmt.Sprintf("%d", i+193), "out of order key: %v", k)
		require.Falsef(t, i == 63 && k != "192", "out of order key: %v", k)
	}

	l.Purge()
	size, err = l.Len()
	require.Nil(t, err)
	require.Equalf(t, size, 0, "bad len: %v", size)
	err = l.Get("200", &v)
	require.Equal(t, err, ErrKeyNotFound, "should contain nothing")

	err = l.Set("201", 301)
	require.Nil(t, err)
	err = l.Get("201", &v)
	require.Nil(t, err)
	require.Equal(t, 301, v)

}

func TestLRUExpire(t *testing.T) {
	l := NewLRU(&LRUOptions{
		Size:                   128,
		DefaultExpiry:          1 * time.Second,
		InvalidateClusterEvent: "",
	})

	l.SetWithDefaultExpiry("1", 1)
	l.SetWithExpiry("3", 3, 0*time.Second)

	time.Sleep(time.Second * 2)

	var r1 int
	err := l.Get("1", &r1)
	require.Equal(t, err, ErrKeyNotFound, "should not exist")

	var r2 int
	err2 := l.Get("3", &r2)
	require.Nil(t, err2, "should exist")
	require.Equal(t, 3, r2)
}

func TestLRUMarshalUnMarshal(t *testing.T) {
	l := NewLRU(&LRUOptions{
		Size:                   1,
		DefaultExpiry:          0,
		InvalidateClusterEvent: "",
	})

	value1 := map[string]interface{}{
		"key1": 1,
		"key2": "value2",
	}
	err := l.Set("test", value1)

	require.Nil(t, err)

	var value2 map[string]interface{}
	err = l.Get("test", &value2)
	require.Nil(t, err)

	v1, ok := value2["key1"].(int64)
	require.True(t, ok, "unable to cast value")
	assert.Equal(t, int64(1), v1)

	v2, ok := value2["key2"].(string)
	require.True(t, ok, "unable to cast value")
	assert.Equal(t, "value2", v2)

	post := model.Post{
		Id:            "id",
		CreateAt:      11111,
		UpdateAt:      11111,
		DeleteAt:      11111,
		EditAt:        111111,
		IsPinned:      true,
		UserId:        "UserId",
		ChannelId:     "ChannelId",
		RootId:        "RootId",
		ParentId:      "ParentId",
		OriginalId:    "OriginalId",
		Message:       "OriginalId",
		MessageSource: "MessageSource",
		Type:          "Type",
		Props: map[string]interface{}{
			"key": "val",
		},
		Hashtags:      "Hashtags",
		Filenames:     []string{"item1", "item2"},
		FileIds:       []string{"item1", "item2"},
		PendingPostId: "PendingPostId",
		HasReactions:  true,
		ReplyCount:    11111,
		Metadata: &model.PostMetadata{
			Embeds: []*model.PostEmbed{
				{
					Type: "Type",
					URL:  "URL",
					Data: "some data",
				},
				{
					Type: "Type 2",
					URL:  "URL 2",
					Data: "some data 2",
				},
			},
			Emojis: []*model.Emoji{
				{
					Id:   "id",
					Name: "name",
				},
			},
			Files: nil,
			Images: map[string]*model.PostImage{
				"key": {
					Width:      1,
					Height:     1,
					Format:     "format",
					FrameCount: 1,
				},
				"key2": {
					Width:      999,
					Height:     888,
					Format:     "format 2",
					FrameCount: 1000,
				},
			},
			Reactions: []*model.Reaction{
				{
					UserId:    "user_id",
					PostId:    "post_id",
					EmojiName: "emoji_name",
					CreateAt:  111,
				},
			},
		},
	}
	err = l.Set("post", post.Clone())
	require.Nil(t, err)

	var p model.Post
	err = l.Get("post", &p)
	require.Nil(t, err)
	require.Equal(t, post.Clone(), p.Clone())
}

func BenchmarkLRU(b *testing.B) {

	value1 := "simplestring"
	b.Run("simple=old", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l := lru.New(1)
			l.Add("test", value1)
			_, ok := l.Get("test")
			require.True(b, ok)
		}
	})

	b.Run("simple=new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l2 := NewLRU(&LRUOptions{
				Size:                   1,
				DefaultExpiry:          0,
				InvalidateClusterEvent: "",
			})
			err := l2.Set("test", value1)
			require.Nil(b, err)

			var val string
			err = l2.Get("test", &val)
			require.Nil(b, err)
		}
	})

	type obj struct {
		Field1 int
		Field2 string
		Field3 struct {
			Field4 int
			Field5 string
		}
		Field6 map[string]string
	}

	value2 := obj{
		1,
		"field2",
		struct {
			Field4 int
			Field5 string
		}{
			6,
			"field5 is a looooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong string",
		},
		map[string]string{
			"key0": "value0",
			"key1": "value value1",
			"key2": "value value value2",
			"key3": "value value value value3",
			"key4": "value value value value value4",
			"key5": "value value value value value value5",
			"key6": "value value value value value value value6",
			"key7": "value value value value value value value value7",
			"key8": "value value value value value value value value value8",
			"key9": "value value value value value value value value value value9",
		},
	}
	b.Run("complex=old", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l := lru.New(1)
			l.Add("test", value2)
			_, ok := l.Get("test")
			require.True(b, ok)
		}
	})
	b.Run("complex=new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l2 := NewLRU(&LRUOptions{
				Size:                   1,
				DefaultExpiry:          0,
				InvalidateClusterEvent: "",
			})
			err := l2.Set("test", value2)
			require.Nil(b, err)

			var val obj
			err = l2.Get("test", &val)
			require.Nil(b, err)
		}
	})

	user := &model.User{
		Id:             "id",
		CreateAt:       11111,
		UpdateAt:       11111,
		DeleteAt:       11111,
		Username:       "username",
		Password:       "password",
		AuthService:    "AuthService",
		AuthData:       nil,
		Email:          "Email",
		EmailVerified:  true,
		Nickname:       "Nickname",
		FirstName:      "FirstName",
		LastName:       "LastName",
		Position:       "Position",
		Roles:          "Roles",
		AllowMarketing: true,
		Props: map[string]string{
			"key0": "value0",
			"key1": "value value1",
			"key2": "value value value2",
			"key3": "value value value value3",
			"key4": "value value value value value4",
			"key5": "value value value value value value5",
			"key6": "value value value value value value value6",
			"key7": "value value value value value value value value7",
			"key8": "value value value value value value value value value8",
			"key9": "value value value value value value value value value value9",
		},
		NotifyProps: map[string]string{
			"key0": "value0",
			"key1": "value value1",
			"key2": "value value value2",
			"key3": "value value value value3",
			"key4": "value value value value value4",
			"key5": "value value value value value value5",
			"key6": "value value value value value value value6",
			"key7": "value value value value value value value value7",
			"key8": "value value value value value value value value value8",
			"key9": "value value value value value value value value value value9",
		},
		LastPasswordUpdate: 111111,
		LastPictureUpdate:  111111,
		FailedAttempts:     111111,
		Locale:             "Locale",
		Timezone: map[string]string{
			"key0": "value0",
			"key1": "value value1",
			"key2": "value value value2",
			"key3": "value value value value3",
			"key4": "value value value value value4",
			"key5": "value value value value value value5",
			"key6": "value value value value value value value6",
			"key7": "value value value value value value value value7",
			"key8": "value value value value value value value value value8",
			"key9": "value value value value value value value value value value9",
		},
		MfaActive:              true,
		MfaSecret:              "MfaSecret",
		LastActivityAt:         111111,
		IsBot:                  true,
		BotDescription:         "field5 is a looooooooooooooooooooooooooooooooooooooooooooooooooooooooooooooong string",
		BotLastIconUpdate:      111111,
		TermsOfServiceId:       "TermsOfServiceId",
		TermsOfServiceCreateAt: 111111,
	}

	b.Run("User=old", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l := lru.New(1)
			l.Add("test", user)
			_, ok := l.Get("test")
			require.True(b, ok)
		}
	})
	b.Run("User=new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l2 := NewLRU(&LRUOptions{
				Size:                   1,
				DefaultExpiry:          0,
				InvalidateClusterEvent: "",
			})
			err := l2.Set("test", user)
			require.Nil(b, err)

			var val model.User
			err = l2.Get("test", &val)
			require.Nil(b, err)
		}
	})

	post := &model.Post{
		Id:            "id",
		CreateAt:      11111,
		UpdateAt:      11111,
		DeleteAt:      11111,
		EditAt:        111111,
		IsPinned:      true,
		UserId:        "UserId",
		ChannelId:     "ChannelId",
		RootId:        "RootId",
		ParentId:      "ParentId",
		OriginalId:    "OriginalId",
		Message:       "OriginalId",
		MessageSource: "MessageSource",
		Type:          "Type",
		Props: map[string]interface{}{
			"key": "val",
		},
		Hashtags:      "Hashtags",
		Filenames:     []string{"item1", "item2"},
		FileIds:       []string{"item1", "item2"},
		PendingPostId: "PendingPostId",
		HasReactions:  true,

		// Transient data populated before sending a post to the client
		ReplyCount: 11111,
		Metadata: &model.PostMetadata{
			Embeds: []*model.PostEmbed{
				{
					Type: "Type",
					URL:  "URL",
					Data: "some data",
				},
				{
					Type: "Type 2",
					URL:  "URL 2",
					Data: "some data 2",
				},
			},
			Emojis: []*model.Emoji{
				{
					Id:   "id",
					Name: "name",
				},
			},
			Files: nil,
			Images: map[string]*model.PostImage{
				"key": {
					Width:      1,
					Height:     1,
					Format:     "format",
					FrameCount: 1,
				},
				"key2": {
					Width:      999,
					Height:     888,
					Format:     "format 2",
					FrameCount: 1000,
				},
			},
			Reactions: []*model.Reaction{},
		},
	}

	b.Run("Post=old", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l := lru.New(1)
			l.Add("test", post)
			_, ok := l.Get("test")
			require.True(b, ok)
		}
	})
	b.Run("Post=new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l2 := NewLRU(&LRUOptions{
				Size:                   1,
				DefaultExpiry:          0,
				InvalidateClusterEvent: "",
			})
			err := l2.Set("test", post)
			require.Nil(b, err)

			var val model.Post
			err = l2.Get("test", &val)
			require.Nil(b, err)
		}
	})

	status := model.Status{
		UserId:         "UserId",
		Status:         "Status",
		Manual:         true,
		LastActivityAt: 111111,
		ActiveChannel:  "ActiveChannel",
	}
	b.Run("Status=old", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l := lru.New(1)
			l.Add("test", status)
			_, ok := l.Get("test")
			require.True(b, ok)
		}
	})
	b.Run("Status=new", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			l2 := NewLRU(&LRUOptions{
				Size:                   1,
				DefaultExpiry:          0,
				InvalidateClusterEvent: "",
			})
			err := l2.Set("test", status)
			require.Nil(b, err)

			var val model.Status
			err = l2.Get("test", &val)
			require.Nil(b, err)
		}
	})
}
