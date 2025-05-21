package clerk_user_center

import (
	"context"
	"fmt"

	"github.com/apache/answer/plugin"
	"github.com/clerk/clerk-sdk-go/v2/clerk"
)

// ClerkUserCenter 实现 plugin.UserCenter 接口
type ClerkUserCenter struct {
	Config *UserCenterConfig
	Client *clerk.Client
}

// init 注册插件
func init() {
	plugin.Register(&ClerkUserCenter{})
}

// Info 返回插件信息
func (c *ClerkUserCenter) Info() plugin.Info {
	return plugin.Info{
		Name:        plugin.MakeTranslator("clerk_user_center"),
		SlugName:    "clerk_user_center",
		Description: plugin.MakeTranslator("Clerk UserCenter 插件，用于替换 Apache Answer 默认认证"),
		Author:      "Sean Chen",
		Version:     "0.1.0",
		Link:        "https://github.com/yourorg/answer-plugins/user-center-clerk",
	}
}

// Description 声明在 Answer 后台显示的入口名称、图标与路由
func (c *ClerkUserCenter) Description() plugin.UserCenterDesc {
	return plugin.UserCenterDesc{
		Name:     plugin.MakeTranslator("Clerk SSO Login"),
		Icon:     `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24"><path d="M12 2C6.5 2 2 6.49 2 12s4.5 10 10 10 10-4.49 10-10S17.5 2 12 2zm0 18c-4.41 0-8-3.59-8-8s3.59-8 8-8 8 3.59 8 8-3.59 8-8 8z"/><path d="M11 6h2v6h-2zm0 8h2v2h-2z"/></svg>`,
		URL:      "/login/clerk",
		Priority: 10,
	}
}

// ControlCenterItems 后台中心显示的快捷链接
func (c *ClerkUserCenter) ControlCenterItems() []plugin.ControlCenter {
	return []plugin.ControlCenter{
		{
			Name:  "Clerk Dashboard",
			Label: "Clerk Dashboard",
			Url:   "https://dashboard.clerk.com",
		},
	}
}

// ensureClient 根据配置初始化 Clerk SDK
func (c *ClerkUserCenter) ensureClient() error {
	if c.Client != nil {
		return nil
	}
	if c.Config == nil || c.Config.SecretKey == "" || c.Config.FrontendAPI == "" {
		return fmt.Errorf("Clerk config is not set")
	}
	c.Client = clerk.NewClient(
		clerk.WithSecretKey(c.Config.SecretKey),
		clerk.WithFrontendAPI(c.Config.FrontendAPI),
	)
	return nil
}

// LoginCallback 处理 Clerk 重定向到 /api/v1/user-center/clerk/login/callback?session_token=...
func (c *ClerkUserCenter) LoginCallback(ctx *plugin.GinContext) (*plugin.UserCenterBasicUserInfo, error) {
	if err := c.ensureClient(); err != nil {
		return nil, err
	}
	sessionToken := ctx.Query("session_token")
	if sessionToken == "" {
		return nil, fmt.Errorf("Lack of Clerk session_token parameter in URL")
	}
	session, err := c.Client.Sessions.GetSessionByToken(ctx, sessionToken)
	if err != nil {
		return nil, fmt.Errorf("Clerk: Invalid session_token: %w", err)
	}
	clerkUser, err := c.Client.Users.User(ctx, session.UserID)
	if err != nil {
		return nil, fmt.Errorf("Clerk: Failed to get User Info: %w", err)
	}
	// 判空处理，避免 PrimaryEmailAddress 为空导致 panic
	email := ""
	if clerkUser.PrimaryEmailAddress != nil {
		email = clerkUser.PrimaryEmailAddress.EmailAddress
	}
	displayName := clerkUser.FirstName
	if clerkUser.LastName != "" {
		displayName = fmt.Sprintf("%s %s", clerkUser.FirstName, clerkUser.LastName)
	}
	userInfo := &plugin.UserCenterBasicUserInfo{
		ExternalID:  clerkUser.ID,
		Username:    clerkUser.Username,
		DisplayName: displayName,
		Email:       email,
		AvatarURL:   clerkUser.ImageURL,
		Roles:       []string{"member"},
	}
	return userInfo, nil
}

// UserInfo 根据 externalID 拉取用户信息
func (c *ClerkUserCenter) UserInfo(externalID string) (*plugin.UserCenterBasicUserInfo, error) {
	if err := c.ensureClient(); err != nil {
		return nil, err
	}
	clerkUser, err := c.Client.Users.User(context.Background(), externalID)
	if err != nil {
		return nil, fmt.Errorf("Clerk: Failed to get User Info: %w", err)
	}
	email := ""
	if clerkUser.PrimaryEmailAddress != nil {
		email = clerkUser.PrimaryEmailAddress.EmailAddress
	}
	displayName := clerkUser.FirstName
	if clerkUser.LastName != "" {
		displayName = fmt.Sprintf("%s %s", clerkUser.FirstName, clerkUser.LastName)
	}
	return &plugin.UserCenterBasicUserInfo{
		ExternalID:  clerkUser.ID,
		Username:    clerkUser.Username,
		DisplayName: displayName,
		Email:       email,
		AvatarURL:   clerkUser.ImageURL,
		Roles:       []string{"member"},
	}, nil
}

// UserStatus 返回用户状态（正常/禁用等）
func (c *ClerkUserCenter) UserStatus(externalID string) plugin.UserStatus {
	// 可根据 Clerk 用户状态进一步完善
	// clerkUser, err := c.Client.Users.User(context.Background(), externalID)
	// if err == nil && clerkUser.Banned { return plugin.UserStatusBanned }
	return plugin.UserStatusNormal
}

// SignUpCallback 如需特殊处理，可与 LoginCallback 相同或略作修改
func (c *ClerkUserCenter) SignUpCallback(ctx *plugin.GinContext) (*plugin.UserCenterBasicUserInfo, error) {
	return c.LoginCallback(ctx)
}

// UserList 可选：拉取用户列表（如 Answer 后台需要展示用户列表时实现）
func (c *ClerkUserCenter) UserList(page, pageSize int) ([]*plugin.UserCenterBasicUserInfo, int, error) {
	if err := c.ensureClient(); err != nil {
		return nil, 0, err
	}
	users, meta, err := c.Client.Users.List(context.Background(), &clerk.UsersListParams{
		Limit:  int64(pageSize),
		Offset: int64((page - 1) * pageSize),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("Clerk: Failed to list users: %w", err)
	}
	result := make([]*plugin.UserCenterBasicUserInfo, 0, len(users))
	for _, u := range users {
		email := ""
		if u.PrimaryEmailAddress != nil {
			email = u.PrimaryEmailAddress.EmailAddress
		}
		displayName := u.FirstName
		if u.LastName != "" {
			displayName = fmt.Sprintf("%s %s", u.FirstName, u.LastName)
		}
		result = append(result, &plugin.UserCenterBasicUserInfo{
			ExternalID:  u.ID,
			Username:    u.Username,
			DisplayName: displayName,
			Email:       email,
			AvatarURL:   u.ImageURL,
			Roles:       []string{"member"},
		})
	}
	total := 0
	if meta != nil {
		total = int(meta.TotalCount)
	}
	return result, total, nil
}

// UserSettings 可选：返回用户中心设置入口
func (c *ClerkUserCenter) UserSettings(externalID string) *plugin.UserCenterSetting {
	return &plugin.UserCenterSetting{
		Name:  "Clerk Settings",
		Label: "Manage my Clerk Account",
		Url:   "https://dashboard.clerk.com/user",
	}
}

// PersonalBranding 可选：返回个人品牌信息
func (c *ClerkUserCenter) PersonalBranding(externalID string) *plugin.UserCenterBranding {
	if err := c.ensureClient(); err != nil {
		return &plugin.UserCenterBranding{}
	}
	clerkUser, err := c.Client.Users.User(context.Background(), externalID)
	if err != nil {
		return &plugin.UserCenterBranding{}
	}
	return &plugin.UserCenterBranding{
		AvatarURL: clerkUser.ImageURL,
		Nickname:  clerkUser.Username,
	}
}
